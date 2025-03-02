terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region = "us-east-1"
}

resource "aws_s3_bucket" "test_bucket" {
  bucket = "kylo-rilo-bucket-us-east-1"

  tags = {
    Env = "Dev"
    IaC = true
  }
}

resource "aws_iam_role" "test_role" {
    name = "test_s3_role"
    assume_role_policy = jsonencode({
        Version = "2012-10-17"
        Statement = [{
            Action = "sts:AssumeRole"
            Effect = "Allow"
            Principal = {
                ExternalId = "c6ff1bd6-8a46-4f0e-b2b5-733ff08db06b"
            }
        }]
    })
    
    tags = {
        Env = "Dev"
        IaC = true
    }
}

resource "aws_iam_role_policy" "test_policy" {
    name = "test_s3_policy"
    role = aws_iam_role.test_role.id
    policy = jsonencode({
        Version = "2012-10-17"
        Statement = [{
            Effect = "Allow"
            Resource = "s3://*"
            Action = [
                "s3:ListObjects",
                "s3:GetObject"
            ]
        }]
    })
}