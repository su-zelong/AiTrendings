docker run -d \
  --name my-postgres \
  --restart=always \
  --user 1000:1000 \
  -e POSTGRES_PASSWORD=mysecretpassword \
  -e TZ=Asia/Shanghai \
  -p 5432:5432 \
  -v /mnt/d/SZL-program/postgres:/var/lib/postgresql/data \
  postgres:latest