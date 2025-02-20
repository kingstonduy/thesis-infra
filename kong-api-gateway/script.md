docker network create kong-net

docker run -d --name kong-database \
--network=kong-net \
-e "POSTGRES_USER=kong" \
-e "POSTGRES_DB=kong" \
-e "POSTGRES_PASSWORD=kongpass" \
postgres:13

docker run --rm --network=kong-net \
-e "KONG_DATABASE=postgres" \
-e "KONG_PG_HOST=kong-database" \
-e "KONG_PG_PASSWORD=kongpass" \
-e "KONG_PASSWORD=test" \
kong/kong-gateway:3.9.0.0 kong migrations bootstrap

docker run -d --name kong-gateway \
--network=kong-net \
-e "KONG_DATABASE=postgres" \
-e "KONG_PG_HOST=kong-database" \
-e "KONG_PG_USER=kong" \
-e "KONG_PG_PASSWORD=kongpass" \
-e "KONG_PROXY_ACCESS_LOG=/dev/stdout" \
-e "KONG_ADMIN_ACCESS_LOG=/dev/stdout" \
-e "KONG_PROXY_ERROR_LOG=/dev/stderr" \
-e "KONG_ADMIN_ERROR_LOG=/dev/stderr" \
-e "KONG_ADMIN_LISTEN=0.0.0.0:8001" \
-e "KONG_ADMIN_GUI_URL=http://34.124.235.0:8002" \
-e KONG_LICENSE_DATA \
-p 8000:8000 \
-p 8443:8443 \
-p 8001:8001 \
-p 8444:8444 \
-p 8002:8002 \
-p 8445:8445 \
-p 8003:8003 \
-p 8004:8004 \
kong/kong-gateway:3.9.0.0

docker kill kong-database
docker container rm kong-database
docker network rm kong-net

# Register Product Service

curl -i -X POST http://localhost:8001/services/ \
 --data "name=product-service" \
 --data "url=http://product-service:7002"

# Register Order Service

curl -i -X POST http://localhost:8001/services/ \
 --data "name=order-service" \
 --data "url=http://order-service:7005"

# Register Cart Service

curl -i -X POST http://localhost:8001/services/ \
 --data "name=cart-service" \
 --data "url=http://cart-service:7003"

# Register User Service

curl -i -X POST http://localhost:8001/services/ \
 --data "name=user-service" \
 --data "url=http://user-service:7001"

# Route for Product Service

curl -i -X POST http://localhost:8001/routes/ \
 --data "paths[]=/product-service" \
 --data "service.name=product-service"

# Route for Order Service

curl -i -X POST http://localhost:8001/routes/ \
 --data "paths[]=/order-service" \
 --data "service.name=order-service"

# Route for Cart Service

curl -i -X POST http://localhost:8001/routes/ \
 --data "paths[]=/cart-service" \
 --data "service.name=cart-service"

# Route for User Service

curl -i -X POST http://localhost:8001/routes/ \
 --data "paths[]=/user-service" \
 --data "service.name=user-service"
