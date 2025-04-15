# Instant Messaging Simulation

## Prerequisites
### Container runtime
Use [crun](https://github.com/containers/crun) to run containers deamon-less for less memory footprint.
```bash
wget https://github.com/containers/crun/releases/download/1.20/crun-1.20-linux-amd64
sudo mv crun-1.20-linux-amd64 /usr/bin/crun
sudo chmod +x /usr/bin/crun
```

Add or replace `daemon.json` to enable use of `crun` runtime.
```bash
mv daemon.json /etc/docker/
```

This repository depends on tshark which can be found in most package managers such as apt. 
```bash
apt install tshark
```
You need to change your permisisons such that you are part of the wireshark group
```bash
sudo usermod -a -G wireshark $USER
```
It is probably not ideal to be part of this group long-term so you might want to remove yourself from the group when you are done with this repositoty

## Commands

### Create certificates
Create the certificates used by clients and server before building
```bash
cd cmd/signal-sim/cert/
./generate_cert.sh
```

### Run signal protocol simulation
Run a simulation with N clients
```bash
make signal
```

### Stop simulation
Stop all running containers
```bash
make stop
```

### Reset
Stop running containers and remove network IMvlan
```bash
make reset
```
