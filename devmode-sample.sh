#! /bin/sh

export CCS_CUSTOMERS="pancake,burger,test"
export CCS_STRIPE_PRIVATE="sk_live_abc123"
export CCS_STRIPE_PUBLIC="pk_live_abc123"

export CCS_HTTP_PORT=":8080"

export CCS_CACHEDIR="." # won't be used in dev mode
export CCS_EMAIL="foo@example.com" # won't be used in dev mode
export CCS_HTTP_ONLY="true"
export CCS_LOCALHOST_OVERRIDE="leakygas.yourdomain.com"

export PANCAKE_HOSTNAME="pancakes.yourdomain.com"
export PANCAKE_PATH="syrup"
export PANCAKE_NAME="Joe's Pancake Shack"
export PANCAKE_STRIPE_CUST="baz"

export BURGER_HOSTNAME="burgerbarn.yourdomain.com"
export BURGER_PATH="ketchup"
export BURGER_NAME="Bob's Burger Barn"
export BURGER_STRIPE_CUST="baz"

export TEST_HOSTNAME="leakygas.yourdomain.com"
export TEST_PATH="nozzle"
export TEST_NAME="Leaky's Gas Station"
export TEST_STRIPE_CUST="quux"
export TEST_STRIPE_PRIVATE="sk_test_xyz789"
export TEST_STRIPE_PUBLIC="pk_test_xyz789"

exec go run -tags=dev .
