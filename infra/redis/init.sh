#!/bin/bash

# Redis initialization script for nullmail development

echo "Initializing Redis for nullmail development..."

# Wait for Redis to be ready
until redis-cli -a dev123 ping; do
    echo "Waiting for Redis to be ready..."
    sleep 1
done

echo "Redis is ready!"

# Set some initial keys for development
redis-cli -a dev123 <<EOF
# Create some sample data for testing
SET nullmail:version "1.0.0-dev"
SET nullmail:env "development"
HSET nullmail:stats emails_sent 0
HSET nullmail:stats emails_received 0
HSET nullmail:stats last_startup "$(date)"

# Create sample emails for testing with proper structure
SET nullmail:email:test-1 '{"id":"test-1","from":"noreply@example.com","subject":"Welcome to NullMail!","body":{"text":"Thank you for trying NullMail. This is your first test email.","raw":"Thank you for trying NullMail. This is your first test email."},"recipients":["test@nullmail.local","demo@nullmail.local"],"received_at":"'$(date -Iseconds)'","read":false,"starred":true,"headers":{},"attachments":[]}'
SET nullmail:email:test-2 '{"id":"test-2","from":"support@company.com","subject":"Account Created","body":{"text":"Your account has been successfully created. You can now start receiving emails.","raw":"Your account has been successfully created. You can now start receiving emails."},"recipients":["test@nullmail.local"],"received_at":"'$(date -d '1 hour ago' -Iseconds)'","read":true,"attachments":[],"headers":{}}'
SET nullmail:email:test-3 '{"id":"test-3","from":"security@alerts.com","subject":"Security Notice","body":{"text":"This is a security notification for your account.","raw":"This is a security notification for your account."},"recipients":["demo@nullmail.local"],"received_at":"'$(date -d '2 hours ago' -Iseconds)'","read":false,"starred":false,"headers":{},"attachments":[]}'

# Add emails to recipient-indexed lists (new approach)
LPUSH emails:test@nullmail.local test-1
LPUSH emails:test@nullmail.local test-2  

LPUSH emails:demo@nullmail.local test-1
LPUSH emails:demo@nullmail.local test-3

# Also add to general email list for backward compatibility
LPUSH nullmail:emails test-1
LPUSH nullmail:emails test-2
LPUSH nullmail:emails test-3

# Set expiration for development data (24 hours)
EXPIRE nullmail:email:test-1 86400
EXPIRE nullmail:email:test-2 86400
EXPIRE nullmail:email:test-3 86400
EXPIRE emails:test@nullmail.local 86400
EXPIRE emails:demo@nullmail.local 86400
EXPIRE nullmail:emails 86400

ECHO "Redis initialized with sample data"
EOF

echo "Redis initialization complete!"