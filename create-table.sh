aws  --endpoint http://localhost:8000 dynamodb create-table --table-name checklist1 \
--attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S AttributeName=completed,AttributeType=S \
--key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE \
--provisioned-throughput \
        ReadCapacityUnits=10,WriteCapacityUnits=5 \
--local-secondary-indexes \
        "[{\"IndexName\": \"completedItemsIndex\",
        \"KeySchema\":[{\"AttributeName\":\"PK\",\"KeyType\":\"HASH\"},
                      {\"AttributeName\":\"completed\",\"KeyType\":\"RANGE\"}],
        \"Projection\":{\"ProjectionType\":\"ALL\"}}]"
