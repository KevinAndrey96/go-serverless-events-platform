#!/usr/bin/env bash
set -euo pipefail

# Configuration
ENDPOINT_URL="${DYNAMO_ENDPOINT:-http://localhost:8000}"
TABLE_NAME="${TABLE_NAME:-EventsPlatformTable}"
REGION="us-east-1"

echo "=========================================================="
echo "Initializing DynamoDB Local Table: ${TABLE_NAME}"
echo "Endpoint: ${ENDPOINT_URL}"
echo "=========================================================="

export AWS_ACCESS_KEY_ID="dummy_key"
export AWS_SECRET_ACCESS_KEY="dummy_secret"

# Check if table already exists
if aws dynamodb describe-table --table-name "${TABLE_NAME}" --endpoint-url "${ENDPOINT_URL}" --region "${REGION}" >/dev/null 2>&1; then
    echo "⚠️ Table '${TABLE_NAME}' already exists. Skipping creation."
else
    echo "==> Creating Single-Table Schema with GSIs (sk-index & eventid-index)..."
    aws dynamodb create-table \
        --table-name "${TABLE_NAME}" \
        --endpoint-url "${ENDPOINT_URL}" \
        --region "${REGION}" \
        --attribute-definitions \
            AttributeName=pk,AttributeType=S \
            AttributeName=sk,AttributeType=S \
            AttributeName=event_id,AttributeType=S \
        --key-schema \
            AttributeName=pk,KeyType=HASH \
            AttributeName=sk,KeyType=RANGE \
        --billing-mode PAY_PER_REQUEST \
        --global-secondary-indexes \
            "[
                {
                    \"IndexName\": \"sk-index\",
                    \"KeySchema\": [
                        {\"AttributeName\": \"sk\", \"KeyType\": \"HASH\"},
                        {\"AttributeName\": \"pk\", \"KeyType\": \"RANGE\"}
                    ],
                    \"Projection\": {\"ProjectionType\": \"ALL\"}
                },
                {
                    \"IndexName\": \"eventid-index\",
                    \"KeySchema\": [
                        {\"AttributeName\": \"event_id\", \"KeyType\": \"HASH\"},
                        {\"AttributeName\": \"sk\", \"KeyType\": \"RANGE\"}
                    ],
                    \"Projection\": {\"ProjectionType\": \"ALL\"}
                }
            ]"
    echo "✅ Table '${TABLE_NAME}' created successfully."
fi

echo "==> Inserting sample development seed fixtures..."

# Seed Event
aws dynamodb put-item \
    --table-name "${TABLE_NAME}" \
    --endpoint-url "${ENDPOINT_URL}" \
    --region "${REGION}" \
    --item '{
        "pk": {"S": "EVENT#1001"},
        "sk": {"S": "METADATA#1001"},
        "event_id": {"S": "1001"},
        "name": {"S": "Cloud Native Architecture Summit 2026"},
        "status": {"S": "ACTIVE"},
        "capacity": {"N": "500"},
        "location": {"S": "Bogota Convention Center"}
    }'

# Seed Attendee / Registration
aws dynamodb put-item \
    --table-name "${TABLE_NAME}" \
    --endpoint-url "${ENDPOINT_URL}" \
    --region "${REGION}" \
    --item '{
        "pk": {"S": "EVENT#1001"},
        "sk": {"S": "USER#501"},
        "event_id": {"S": "1001"},
        "user_id": {"S": "501"},
        "user_name": {"S": "Kevin Andrey Herrera Silva"},
        "user_email": {"S": "kahs_kevin@hotmail.com"},
        "payment_status": {"S": "PAID"},
        "referral_code": {"S": "REF-SUMMIT-2026"}
    }'

echo "✅ Seed data populated successfully!"
echo "   - Event: Cloud Native Architecture Summit 2026 (EVENT#1001)"
echo "   - Attendee: Kevin Andrey Herrera Silva (USER#501) [PAID]"
echo "=========================================================="
echo "You can inspect the table in DynamoDB Admin at http://localhost:8001"
