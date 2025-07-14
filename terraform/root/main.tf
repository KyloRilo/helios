terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
  assume_role {
    role_arn = var.helios_role_config[terraform.workspace].role_arn
    external_id = var.helios_role_config[terraform.workspace].external_id
    session_name = "helios-${terraform.workspace}-tf-session"
  } 
}

resource "aws_ecr_repository" "helios" {
  name = "helios-${terraform.workspace}"
  image_tag_mutability = "IMMUTABLE"
  tags = {
    Env = terraform.workspace
    IaC = true
  }
}