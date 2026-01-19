terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

variable "region" {
  default = "us-west-2"
}

resource "aws_instance" "app" {
  ami           = "ami-12345678"
  instance_type = "t3.micro"

  tags = {
    Name = "app-server"
  }
}
