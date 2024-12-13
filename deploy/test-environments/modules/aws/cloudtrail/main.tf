locals {
  s3_bucket_name = "tf-test-envs-cloudtrail-logs"
  common_tags = {
    division = "${var.division}"
    org      = "${var.org}"
    team     = "${var.team}"
    project  = "${var.project}"
    owner    = "${var.owner}"
  }
}

resource "aws_s3_bucket" "cloudtrail" {
  bucket        = var.s3_bucket_name
  force_destroy = true
  tags          = local.common_tags
}

resource "aws_kms_key" "cloudtrail" {
  description = "KMS key for CloudTrail logs encryption"

  policy = jsonencode({
    Version = "2012-10-17",
    Id      = "key-default-1",
    Statement : [
      {
        Sid    = "Enable IAM User Permissions",
        Effect = "Allow",
        Principal = {
          AWS = "arn:aws:iam::${var.aws_account_id}:root"
        },
        Action   = "kms:*",
        Resource = "*"
      },
      {
        Sid    = "Allow CloudTrail to encrypt logs",
        Effect = "Allow",
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        },
        Action = [
          "kms:GenerateDataKey*",
          "kms:Decrypt"
        ],
        Resource = "*",
        Condition = {
          StringLike = {
            "kms:EncryptionContext:aws:cloudtrail:arn" = "arn:aws:cloudtrail:*:${var.aws_account_id}:trail/*"
          }
        }
      },
      {
        Sid    = "Allow CloudTrail to describe key",
        Effect = "Allow",
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        },
        Action   = "kms:DescribeKey",
        Resource = "*"
      }
    ]
  })

  tags = local.common_tags
}

resource "aws_kms_alias" "cloudtrail" {
  name          = "alias/${var.kms_alias_name}"
  target_key_id = aws_kms_key.cloudtrail.id
}

resource "aws_cloudtrail" "main" {
  name                          = var.cloudtrail_name
  s3_bucket_name                = aws_s3_bucket.cloudtrail.bucket
  include_global_service_events = true
  is_multi_region_trail         = true
  enable_log_file_validation    = true
  kms_key_id                    = aws_kms_key.cloudtrail.arn

  insight_selector {
    insight_type = "ApiCallRateInsight"
  }

  insight_selector {
    insight_type = "ApiErrorRateInsight"
  }

  advanced_event_selector {
    name = "Log management events"

    field_selector {
      field  = "eventCategory"
      equals = ["Management"]
    }
  }

  advanced_event_selector {
    name = "AWS App Config"

    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }
    field_selector {
      field  = "resources.type"
      equals = ["AWS::AppConfig::Configuration"]
    }
  }

  advanced_event_selector {
    name = "S3 Object Data"

    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }
    field_selector {
      field  = "resources.type"
      equals = ["AWS::S3::Object"]
    }
  }

  advanced_event_selector {
    name = "DynamoDB Table Data"

    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }
    field_selector {
      field  = "resources.type"
      equals = ["AWS::DynamoDB::Table"]
    }
  }

  tags = local.common_tags
}

resource "aws_s3_bucket_policy" "cloudtrail" {
  bucket = aws_s3_bucket.cloudtrail.bucket

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = "arn:aws:s3:::${aws_s3_bucket.cloudtrail.bucket}"
      },
      {
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "arn:aws:s3:::${aws_s3_bucket.cloudtrail.bucket}/AWSLogs/${var.aws_account_id}/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      }
    ]
  })
}
