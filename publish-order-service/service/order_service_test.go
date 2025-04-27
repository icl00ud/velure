// service/order_service_test.go
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/icl00ud/publish-order-service/pkg/model"
)

type fakeRepo struct {
	saved     model.Order
	saveErr   error
	findOrder model.Order
	findErr   error
	saveCalls int
	findCalls int
}

func (f *fakeRepo) Save(ctx context.Context, o model.Order) error {
	f.saveCalls++
	f.saved = o
	return f.saveErr
}

func (f *fakeRepo) Find(ctx context.Context, id string) (model.Order, error) {
	f.findCalls++
	return f.findOrder, f.findErr
}

func TestOrderService_Create(t *testing.T) {
	pricing := NewPricingCalculator()

	tests := []struct {
		name      string
		items     []model.CartItem
		repoErr   error
		wantErr   error
		wantTotal int
	}{
		{
			name:    "no items",
			items:   nil,
			wantErr: ErrNoItems,
		},
		{
			name:    "repo save error",
			items:   []model.CartItem{{ProductID: "p", Name: "n", Quantity: 1, Price: 5.0}},
			repoErr: errors.New("save failed"),
			wantErr: errors.New("save failed"),
		},
		{
			name:      "success",
			items:     []model.CartItem{{ProductID: "p1", Name: "n1", Quantity: 2, Price: 10.5}},
			wantErr:   nil,
			wantTotal: 21, // 2 * 10.5 = 21
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{saveErr: tt.repoErr}
			svc := NewOrderService(repo, pricing)
			ctx := context.Background()

			got, err := svc.Create(ctx, tt.items)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}
			if got.Total != tt.wantTotal {
				t.Errorf("Create() total = %d, want %d", got.Total, tt.wantTotal)
			}
			if repo.saveCalls != 1 {
				t.Errorf("Save called %d times, want 1", repo.saveCalls)
			}
			if got.ID == "" {
				t.Error("Create() generated empty ID")
			}
		})
	}
}

func TestOrderService_UpdateStatus(t *testing.T) {
	pricing := NewPricingCalculator()
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name      string
		initial   model.Order
		findErr   error
		saveErr   error
		newStatus string
		wantErr   error
	}{
		{
			name:    "find error",
			findErr: errors.New("not found"),
			wantErr: errors.New("not found"),
		},
		{
			name:      "save error",
			initial:   model.Order{ID: "id1", Status: model.OrderCreated, CreatedAt: now, UpdatedAt: now},
			newStatus: model.OrderProcessing,
			saveErr:   errors.New("save failed"),
			wantErr:   errors.New("save failed"),
		},
		{
			name:      "success",
			initial:   model.Order{ID: "id2", Status: model.OrderCreated, CreatedAt: now, UpdatedAt: now},
			newStatus: model.OrderProcessed,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{
				findOrder: tt.initial,
				findErr:   tt.findErr,
				saveErr:   tt.saveErr,
			}
			svc := NewOrderService(repo, pricing)
			ctx := context.Background()

			got, err := svc.UpdateStatus(ctx, tt.initial.ID, tt.newStatus)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("UpdateStatus() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("UpdateStatus() unexpected error: %v", err)
			}
			if got.Status != tt.newStatus {
				t.Errorf("UpdateStatus() status = %q, want %q", got.Status, tt.newStatus)
			}
			if repo.findCalls != 1 {
				t.Errorf("Find called %d times, want 1", repo.findCalls)
			}
			if repo.saveCalls != 1 {
				t.Errorf("Save called %d times, want 1", repo.saveCalls)
			}
			if got.UpdatedAt.Equal(tt.initial.UpdatedAt) {
				t.Error("UpdateStatus() did not update UpdatedAt")
			}
		})
	}
}
