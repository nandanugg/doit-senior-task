# Terraform Infrastructure for URL Shortener

This Terraform configuration deploys the URL Shortener application to AWS using:
- **ECR** - Elastic Container Registry for Docker images
- **ECS Fargate** - Serverless container orchestration
- **RDS PostgreSQL** - Managed database service
- **ALB** - Application Load Balancer
- **VPC** - Isolated network with public/private subnets
- **IAM** - Least privilege roles and policies

## Prerequisites

1. **AWS CLI** configured with appropriate credentials
2. **Terraform** >= 1.0
3. **Docker** for building and pushing images

## Quick Start

### 1. Initialize Terraform

```bash
cd terraform
terraform init
```

### 2. Review and Customize Variables

Create a `terraform.tfvars` file (optional):

```hcl
aws_region     = "us-east-1"
environment    = "dev"
desired_count  = 2
cpu            = 256
memory         = 512
```

### 3. Plan Infrastructure

```bash
terraform plan
```

### 4. Deploy Infrastructure

```bash
terraform apply
```

This will create all necessary AWS resources. Review the plan and type `yes` to confirm.

### 5. Get Outputs

```bash
terraform output
```

Important outputs:
- `ecr_repository_url` - Push your Docker images here
- `alb_url` - Access your application
- `cicd_credentials_secret_arn` - CI/CD user credentials (in Secrets Manager)
- `db_secret_arn` - Database credentials (in Secrets Manager)

## Deploy Application

### 1. Build and Push Docker Image

```bash
# Get ECR repository URL
ECR_URL=$(terraform output -raw ecr_repository_url)
AWS_REGION=$(terraform output -raw aws_region || echo "us-east-1")

# Login to ECR
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $ECR_URL

# Build image (from project root)
cd ..
docker build -t urlshortener:latest .

# Tag and push
docker tag urlshortener:latest $ECR_URL:latest
docker push $ECR_URL:latest
```

### 2. Update ECS Service

The ECS service will automatically deploy the new image:

```bash
cd terraform
terraform apply -var="ecr_image_tag=latest"
```

Or update the service manually:

```bash
aws ecs update-service \
  --cluster $(terraform output -raw ecs_cluster_name) \
  --service $(terraform output -raw ecs_service_name) \
  --force-new-deployment
```

### 3. Access Application

```bash
curl $(terraform output -raw alb_url)/healthz
```

## CI/CD Setup

### Get CI/CD Credentials

```bash
# Get the secret ARN
SECRET_ARN=$(terraform output -raw cicd_credentials_secret_arn)

# Retrieve credentials
aws secretsmanager get-secret-value --secret-id $SECRET_ARN --query SecretString --output text | jq .
```

### GitHub Actions Example

Add these secrets to your GitHub repository:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_REGION`
- `ECR_REPOSITORY`

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ secrets.AWS_REGION }}

      - name: Login to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v1

      - name: Build, tag, and push image
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          ECR_REPOSITORY: ${{ secrets.ECR_REPOSITORY }}
          IMAGE_TAG: ${{ github.sha }}
        run: |
          docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG .
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
          docker tag $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG $ECR_REGISTRY/$ECR_REPOSITORY:latest
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:latest

      - name: Update ECS service
        run: |
          aws ecs update-service \
            --cluster urlshortener-dev \
            --service urlshortener-dev \
            --force-new-deployment
```

## Database Migrations

### Run Migrations

The database is automatically created, but you need to run migrations:

```bash
# Get database connection details from Secrets Manager
SECRET_ARN=$(terraform output -raw db_secret_arn)
DB_URL=$(aws secretsmanager get-secret-value --secret-id $SECRET_ARN --query SecretString --output text | jq -r .url)

# Run migrations (from project root)
cd ..
migrate -path ./migrations -database "$DB_URL" up
```

### Via ECS Exec (Alternative)

```bash
# Enable ECS Exec (already enabled in terraform)
TASK_ID=$(aws ecs list-tasks \
  --cluster $(terraform output -raw ecs_cluster_name) \
  --service-name $(terraform output -raw ecs_service_name) \
  --query 'taskArns[0]' --output text | cut -d'/' -f3)

# Connect to container
aws ecs execute-command \
  --cluster $(terraform output -raw ecs_cluster_name) \
  --task $TASK_ID \
  --container urlshortener \
  --interactive \
  --command "/bin/sh"
```

## Monitoring and Logs

### View Logs

```bash
# Get log group
LOG_GROUP=$(terraform output -raw cloudwatch_log_group)

# Stream logs
aws logs tail $LOG_GROUP --follow
```

### CloudWatch Insights

```bash
aws logs start-query \
  --log-group-name $(terraform output -raw cloudwatch_log_group) \
  --start-time $(date -u -d '1 hour ago' +%s) \
  --end-time $(date -u +%s) \
  --query-string 'fields @timestamp, @message | sort @timestamp desc | limit 20'
```

## Auto Scaling

The ECS service is configured with auto-scaling based on:
- **CPU Utilization**: Target 70%
- **Memory Utilization**: Target 80%
- **Min Tasks**: 2 (configurable via `desired_count`)
- **Max Tasks**: 10

## Security

### IAM Roles and Policies (Principle of Least Privilege)

1. **ECS Task Execution Role**
   - Pull images from ECR
   - Write logs to CloudWatch
   - Read secrets from Secrets Manager

2. **ECS Task Role**
   - Read database credentials from Secrets Manager

3. **CI/CD User**
   - Push images to ECR
   - Update ECS service
   - Register new task definitions

### Secrets Management

All sensitive data is stored in AWS Secrets Manager:
- Database credentials
- CI/CD user credentials

## Cost Optimization

- **Fargate Spot**: Mixed capacity provider strategy
- **RDS**: t3.micro instance (adjustable)
- **NAT Gateways**: 2 for HA (can reduce to 1 for dev)
- **Log Retention**: 7 days (adjustable)

To reduce costs for development:

```hcl
# terraform.tfvars
desired_count = 1
cpu = 256
memory = 512
db_instance_class = "db.t3.micro"
```

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

**Warning**: This will delete all data including the database.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                        Internet                         │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
            ┌────────────────┐
            │   Application  │
            │ Load Balancer  │
            └────────┬───────┘
                     │
        ┌────────────┴────────────┐
        │                         │
        ▼                         ▼
┌───────────────┐         ┌───────────────┐
│ Public Subnet │         │ Public Subnet │
│     (AZ-1)    │         │     (AZ-2)    │
└───────┬───────┘         └───────┬───────┘
        │                         │
        │ NAT Gateway             │ NAT Gateway
        │                         │
        ▼                         ▼
┌───────────────┐         ┌───────────────┐
│Private Subnet │         │Private Subnet │
│     (AZ-1)    │         │     (AZ-2)    │
│               │         │               │
│  ┌─────────┐  │         │  ┌─────────┐  │
│  │   ECS   │  │         │  │   ECS   │  │
│  │ Fargate │  │         │  │ Fargate │  │
│  │  Tasks  │  │         │  │  Tasks  │  │
│  └────┬────┘  │         │  └────┬────┘  │
│       │       │         │       │       │
│       └───────┴─────────┴───────┘       │
│                   │                     │
│                   ▼                     │
│            ┌──────────────┐             │
│            │     RDS      │             │
│            │  PostgreSQL  │             │
│            └──────────────┘             │
└─────────────────────────────────────────┘

┌──────────────┐      ┌──────────────┐
│     ECR      │      │   Secrets    │
│  Repository  │      │   Manager    │
└──────────────┘      └──────────────┘
```

## Troubleshooting

### Tasks Not Starting

```bash
# Check service events
aws ecs describe-services \
  --cluster $(terraform output -raw ecs_cluster_name) \
  --services $(terraform output -raw ecs_service_name) \
  --query 'services[0].events[:5]'
```

### Health Check Failures

```bash
# Check target health
aws elbv2 describe-target-health \
  --target-group-arn $(terraform output -raw target_group_arn)
```

### Database Connection Issues

Verify security group rules allow traffic from ECS tasks to RDS on port 5432.

## Support

For issues or questions, please refer to the project documentation or create an issue in the repository.
