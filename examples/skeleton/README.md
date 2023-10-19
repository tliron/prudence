Prudence Example: Skeleton
==========================

    curl "https://localhost:8081/app1/resource1/hello" --location \
    --insecure

    curl "https://localhost/app1/resource1/hello" --location \
    --cacert examples/skeleton/secret/crt.pem \
    --connect-to localhost:443:localhost:8081
