import { Routes } from '@angular/router';

import { HomeComponent } from './core/components/home/home.component';
import { PageNotFoundComponent } from './core/components/page-not-found/page-not-found.component';
import { LoginComponent } from './core/components/login/login.component';
import { CategoryProductsComponent } from './core/components/category-products/category-products.component';

const categoryData = [
  { path: 'shoes', data: { category: 'shoes' } },
  { path: 'shirts', data: { category: 'shirts' } },
];

export const routes: Routes = [
  { path: '', redirectTo: 'home', pathMatch: 'full' },
  { path: 'home', component: HomeComponent },
  { path: 'login', component: LoginComponent },
  ...categoryData.map((category) => ({
    ...category,
    component: CategoryProductsComponent,
  })),
  { path: '**', component: PageNotFoundComponent },
];