#!/bin/bash


# APPLICATION SPECS
export PORT=8081

# COUCHBASE CONNECTION
export COUCHBASE_CONNECTION_URI=localhost
export COUCHBASE_DEFAULT_BUCKET=testbucket
export COUCHBASE_USERNAME=Administrator
export COUCHBASE_PASSWORD=password

# JWT
export JWT_SECRET_KEY=wKe53gwBst34y93
export JWT_SECRET_ISSUER=auth.loyyal.net
export JWT_TOKEN_VALIDITY=1

# NATS


# JAEGER

# SMTP
export SMTP_EMAIL_FROM=no-reply@loyyal.net
export SMTP_EMAIL_REPLY_TO=support@loyyal.net
export SMTP_EMAIL_HOST=smtp.sendgrid.net
export SMTP_EMAIL_PROTOCOL=587
export SMTP_EMAIL_USERNAME=apikey
export SMTP_EMAIL_PASSWORD=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


go run main.go