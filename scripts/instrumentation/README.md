# Instrumentation Demo

## First time

- Run the Sidecar at http://localhost:8080
- `docker-compose up -d`.
- Go to http://localhost:3000 and login with username `admin` and password `admin`.
- Go to http://localhost:3000/connections/datasources/new and add a Prometheus data source with URL http://prometheus:9090.
- Go to http://localhost:3000/dashboard/new and import the `dashboard.json` file included here.

## Updating dashboard.json

- Click **Share** from the dashboard, then **Export**, making sure to tick **Export for sharing externally**.
- Click **View JSON** and copy-paste the result into `dashboard.json`.
