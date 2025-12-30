#!/bin/bash

# URL of the ingest API
URL="http://localhost:8080/api/v1/ingest"

echo "Starting log generation..."
echo "Target: $URL"
echo "Sending 100 logs with 2s delay..."

for i in {1..100}; do
  # Randomly select a scenario to ensure semantic consistency
  SCENARIO=$((RANDOM % 4))

  case $SCENARIO in
    0)
      SERVICE="auth-service"
      # Auth messages
      RAND_MSG=$((RANDOM % 3))
      if [ $RAND_MSG -eq 0 ]; then
        LEVEL="info"
        MESSAGE="User logged in successfully"
      elif [ $RAND_MSG -eq 1 ]; then
        LEVEL="error"
        MESSAGE="User not found"
      else
        LEVEL="error"
        MESSAGE="Invalid credentials"
      fi
      ;;
    1)
      SERVICE="payment-gateway"
      # Payment messages
      RAND_MSG=$((RANDOM % 3))
      if [ $RAND_MSG -eq 0 ]; then
        LEVEL="info"
        MESSAGE="Payment processed"
      elif [ $RAND_MSG -eq 1 ]; then
        LEVEL="error"
        MESSAGE="Payment declined"
      else
        LEVEL="error"
        MESSAGE="Gateway timeout"
      fi
      ;;
    2)
      SERVICE="user-profile"
      # Profile messages
      RAND_MSG=$((RANDOM % 3))
      if [ $RAND_MSG -eq 0 ]; then
        LEVEL="info"
        MESSAGE="Profile updated"
      elif [ $RAND_MSG -eq 1 ]; then
        LEVEL="error"
        MESSAGE="Database connection failed"
      else
        LEVEL="info"
        MESSAGE="Avatar uploaded"
      fi
      ;;
    3)
      SERVICE="inventory-manager"
      # Inventory messages
      RAND_MSG=$((RANDOM % 3))
      if [ $RAND_MSG -eq 0 ]; then
        LEVEL="info"
        MESSAGE="Inventory checked"
      elif [ $RAND_MSG -eq 1 ]; then
        LEVEL="error"
        MESSAGE="Inventory sync error"
      else
        LEVEL="info"
        MESSAGE="Stock updated"
      fi
      ;;
  esac

  # Construct JSON payload
  PAYLOAD="{\"level\": \"$LEVEL\", \"message\": \"$MESSAGE\", \"service\": \"$SERVICE\", \"iteration\": \"$i\"}"

  # Send request
  echo "[$(date +'%T')] Sending $LEVEL log ($SERVICE): $MESSAGE"
  curl -s -X POST "$URL" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD" > /dev/null

  # Wait 2 seconds
  sleep 2
done

echo "Log generation complete."
