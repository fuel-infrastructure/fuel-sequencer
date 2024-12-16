# Instrumentation Demo

## Assumptions

- You are running Docker 20.10.0+ (on Linux) or Docker 18.0.3+ (on Mac), otherwise `host.docker.internal` which is used throughout does not work.
- A sidecar instance is running and will be reachable from the instrumentation running on Docker.

## Running on Linux

If you're on Linux and the Sidecar is running on your host machine, you might run into connectivity issues between Prometheus and the Sidecar.

Workarounds include:

- Run the Prometheus and Grafana setup locally instead of following the steps in this document.
- Whitelist Docker containers in UFW to ensure that they can access services on your host machine.

If you want to whitelist Docker containers in UFW (Uncomplicated Firewall) on Linux:

1. Determine the Docker bridge network: `ip addr show docker0`.
   - This should show something like `inet 172.17.0.1/16 brd 172.17.255.255 scope global docker0`.
   - Note down the subnet, which is `172.17.0.1/16` in the above example.
2. Whitelist the Docker subnet in UFW: `sudo ufw allow from 172.17.0.0/16 to any port 8081` (replace the subnet accordingly).
3. Reload and enable UFW: `sudo ufw reload; sudo ufw enable`.
4. Verify that the whitelisting was added by running `sudo ufw status`.

## First time running

- Depending on what you want to monitor:
  - Run the Sidecar with Prometheus enabled at http://localhost:8081 (or reconfigure `prometheus.yml` accordingly).
  - Run the Sequencer with Prometheus enabled at http://localhost:26660 (or reconfigure `prometheus.yml` accordingly).
- `docker-compose up -d`.
- Go to http://localhost:9000/targets (Prometheus) and ensure that the metrics are being successfully scraped.
- Go to http://localhost:3000 (Grafana) and login with username `admin` and password `admin`.
- Go to http://localhost:3000/connections/datasources/new and add a Prometheus data source with URL http://prometheus:9090 and scrape interval set to **1s**, for finer grain data.
- Go to http://localhost:3000/dashboard/new and import the relevant `dashboard.json` file included here.
- The dashboards assume a job name satisfying the regex `.*fuel.*`. If this is not the case, change the job name from `prometheus.yml` or insert the job name manually in the job field.

## Updating dashboard.json

- Click **Share** from the dashboard, then **Export**, making sure to tick **Export for sharing externally**.
- Click **View JSON** and copy-paste the result into `dashboard.json`.
- If `__inputs.name` was modified, replace all instances `__inputs.name` throughout the file with `DS_PROMETHEUS`. Example: if this is now `DS_PROMETHEUS_SOMETHING`, find and replace this with `DS_PROMETHEUS` in the entire the file.
- Clear `__inputs` entirely by setting `"__inputs": []`, assuming it contains just the `DS_PROMETHEUS` entry. This is because we want the data source to be templated (i.e. drop-down), not a hard-coded input.
