#!/bin/bash

# Force VPC cleanup script
# This script handles the common issue where VPC deletion fails due to lingering dependencies

set -e

VPC_ID="vpc-02228fe5b01d49614"
REGION="us-east-1"

echo "=== VPC Cleanup Script ==="
echo "VPC ID: $VPC_ID"
echo "Region: $REGION"
echo ""

# Function to check for network interfaces
check_enis() {
    echo "Checking for network interfaces..."
    ENIS=$(aws ec2 describe-network-interfaces \
        --region $REGION \
        --filters "Name=vpc-id,Values=$VPC_ID" \
        --query 'NetworkInterfaces[*].[NetworkInterfaceId,Description,Status]' \
        --output text 2>/dev/null || echo "")

    if [ -z "$ENIS" ]; then
        echo "✓ No network interfaces found"
        return 0
    else
        echo "⚠ Found network interfaces:"
        echo "$ENIS"
        return 1
    fi
}

# Function to delete non-default security groups
cleanup_security_groups() {
    echo ""
    echo "Cleaning up security groups..."
    SGs=$(aws ec2 describe-security-groups \
        --region $REGION \
        --filters "Name=vpc-id,Values=$VPC_ID" \
        --query 'SecurityGroups[?GroupName!=`default`].GroupId' \
        --output text 2>/dev/null || echo "")

    if [ -n "$SGs" ]; then
        for sg in $SGs; do
            echo "Attempting to delete security group: $sg"
            aws ec2 delete-security-group --region $REGION --group-id $sg 2>/dev/null || echo "  Could not delete $sg (may have dependencies)"
        done
    else
        echo "✓ No non-default security groups to clean up"
    fi
}

# Function to detach and delete internet gateways
cleanup_igws() {
    echo ""
    echo "Cleaning up internet gateways..."
    IGWs=$(aws ec2 describe-internet-gateways \
        --region $REGION \
        --filters "Name=attachment.vpc-id,Values=$VPC_ID" \
        --query 'InternetGateways[*].InternetGatewayId' \
        --output text 2>/dev/null || echo "")

    if [ -n "$IGWs" ]; then
        for igw in $IGWs; do
            echo "Detaching and deleting IGW: $igw"
            aws ec2 detach-internet-gateway --region $REGION --internet-gateway-id $igw --vpc-id $VPC_ID 2>/dev/null || true
            aws ec2 delete-internet-gateway --region $REGION --internet-gateway-id $igw 2>/dev/null || true
        done
    else
        echo "✓ No internet gateways to clean up"
    fi
}

# Function to delete subnets
cleanup_subnets() {
    echo ""
    echo "Cleaning up subnets..."
    SUBNETS=$(aws ec2 describe-subnets \
        --region $REGION \
        --filters "Name=vpc-id,Values=$VPC_ID" \
        --query 'Subnets[*].SubnetId' \
        --output text 2>/dev/null || echo "")

    if [ -n "$SUBNETS" ]; then
        for subnet in $SUBNETS; do
            echo "Deleting subnet: $subnet"
            aws ec2 delete-subnet --region $REGION --subnet-id $subnet 2>/dev/null || echo "  Could not delete $subnet"
        done
    else
        echo "✓ No subnets to clean up"
    fi
}

# Function to delete route tables
cleanup_route_tables() {
    echo ""
    echo "Cleaning up route tables..."
    RTBs=$(aws ec2 describe-route-tables \
        --region $REGION \
        --filters "Name=vpc-id,Values=$VPC_ID" \
        --query 'RouteTables[?Associations[0].Main!=`true`].RouteTableId' \
        --output text 2>/dev/null || echo "")

    if [ -n "$RTBs" ]; then
        for rtb in $RTBs; do
            echo "Deleting route table: $rtb"
            aws ec2 delete-route-table --region $REGION --route-table-id $rtb 2>/dev/null || echo "  Could not delete $rtb"
        done
    else
        echo "✓ No non-main route tables to clean up"
    fi
}

# Main cleanup process
echo "Starting cleanup process..."
echo ""

# Try to clean up resources
cleanup_igws
cleanup_route_tables
cleanup_subnets
cleanup_security_groups

# Wait for ENIs to be released
echo ""
echo "Waiting for network interfaces to be released..."
MAX_ATTEMPTS=30
ATTEMPT=0

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    if check_enis; then
        echo ""
        echo "✓ All dependencies cleared!"
        break
    fi

    ATTEMPT=$((ATTEMPT + 1))
    echo "  Waiting... (attempt $ATTEMPT/$MAX_ATTEMPTS)"
    sleep 10
done

if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
    echo ""
    echo "⚠ Warning: Max wait time reached. Some ENIs may still exist."
    echo "You may need to manually check the AWS console."
fi

# Now try terraform destroy
echo ""
echo "=== Running Terraform Destroy ==="
cd /Users/icl00ud/repos/velure/infrastructure/terraform
terraform destroy -auto-approve

echo ""
echo "✓ Cleanup complete!"
