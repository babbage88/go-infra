# go-infra - Intro
This repo is currently being used by me as a self-hosted WebAPI to automate various infrastructure related tasks, like creatting DNS records or requesting certificates from Letsencrypt. 
At the moment, I'm not really trying to build this with the intention of anyone ever really using it. I started this to get better at Golang and general infrastructure automation. 

But, I try to comment or make README's whenever I get stuck on something and find that the solutions ends up not being well documented elsewhere. I know I've come accross some random personal repo while 
searching how to do something that's more of a niche topic or tech. So, if you look at the PR history, you'll see me summarizing my bigger changes into the void.  

I know theres probably much better tools to accomplish some of the thing I do in this repo, but I'm basically just doing everything for fun/learning. I tend to try and avoid non-standard or 3rd party libraries as much as I can or until it stops being enjoyable trying to build it myself. 

## Architecture
I'm not the type of guy who generally feels inclined to have a deep discussion on the differences between Vertical Slice vs N-Tier, or best patteren to use for x problem. Luckily in Golang, those things are less relevant/important.

All that being said, I do like to keep code organized in a logical way. Nothing is set it stone, but I like keep everything organized by it's technical area or concern. For example, anything tbat interacts with the database or contains a struct that represents a databse object, goes into a "database" directory at the root. The webapi directory in broken up into the api_handler and api_server packages, obviously api_handler consists of the handler functions and api_server are the functions starts the Mux server and registers the various handlers. The webutils package is anything called by a handler.

I'm sure the rest is pretty straight forward. At the time of writing, this is the current structure, althought this can and likely will change a bit. But the general oranization strategy won't.

```
❯ tree -L 2 --dirsfirst                                                                                                                                                              ─╯
.
├── auth
│   ├── hashing
│   └── tokens
├── cloud_providers
│   └── cloudflare
├── compose
│   ├── compose.dev.yaml
│   ├── compose.orig.yaml
│   └── compose.test.db.yaml
├── database
│   ├── infra_db
│   ├── migrate
│   └── models
├── utils
│   ├── docker_helper
│   ├── env_helper
│   ├── logger
│   ├── test
│   └── type_helper
├── webapi
│   ├── api_handlers
│   └── api_server
├── webutils
│   └── certhandler
├── compose.vps.yaml
├── compose.yaml
├── devtest.sh
├── Docker.env
├── Dockerfile
├── Dockerfile.dev
├── Dockerfile.test
├── do.vps.env
├── exampledot.env
├── go.mod
├── go.sum
├── main
└── main.go

21 directories, 16 files

```


Also, I'm almost positive that there's a way to have more than package in a directory without the compiler yelling at me. I think the each have to have their own main() or something simple. But, for the moment this is working well enough. Although I do want to go back and fix that soon. I don't want multiple modules, just different packages, so that I could put all of the api_*.go files under webapi. 

## Infrastructure 

Maybe it's my background in Ops, but I think implementation details are important in the overall design of a peice of software. Some are more relevant than others obviously, but where and how something is intended to be ran will always end up becoming important at some point or another. 

I've designed everything to run primarily in a self-hosted/on-prem context primarily, with the ability to be ran in a standby fashion on a VPS that's likely to be in a private VPC. 

### Virtualization and Containers

I currently run my Homelab with a three node Proxmox cluster being Hypervisor. Nothing fancy or rack mounted, just some older Dell Optiplexes and mini HP Elitedesks, with a DIY Truenas inside a cheapo ATX case. I've thought about test driving XCP-ng, but never have gotten around to it. 

At the moment, I run a Docker Swarm cluster for anything that can be containerized, so I lean on Swarm secrets pretty heavily atm. I just run a 3-4 node cluster where the hosts share a floating IP address via VRRP/Keepalived. All running on Ubuntu 22.04 Virtual Machines. It's worked pretty well for me so far without any big bugs or frustrations. The main ones I can remember were releated to iptables/NAT related issues where a host would stop forwarding ingress traffic. But, that was only once or twice when I first set it up and fixed by a reboot. I have each node running as a Swarm Manager to avoid failover/HA issues if my one Manager VM went down. 

Things I don't contianerize fully and run at least one VM or LXC instance for: 
- Database
- Logging (Elk Stack)
- DNS

I want to move everything that my wife wouldn't notice went down to a K3s/Kubernetes cluster, but before I do that, I want to get my IaC in fighting condition. I want to be able to run a `terraform apply` and have a threee node cluster boot up fully configured and ready to take workloads, I'm probably 15-25% of the way there at the moment. I have more experience with Docker/Compose/Swarm than I do with Kubernettes. 

### Operating Systems

I don't currently run any Windows Servers in my homelab. I used to run a AD Domain and have the DCs doing DNS. But, I'm not about to pay for a Windows Standard License and I got tired of Plex breaking because my Windows Server Eval limit and the DCs shutting themselves down. So a couple years ago I nuked all my DCs and killed the domain. So I run 100% on Linux now. At some point I'll spin up a domain for testing purposes. But, in general, everything in this repo assumes it's running in a container or *nix OS. I do run Windows 11 and do development on my laptop, but I'm mostly just connecting over ssh to 1 of 2 ubuntu server VMs in VS Code. I still run the full Visual Studio for .Net development sometimes, but I tend to get my fill of that at work. 

### Databases

For this project specificly, I use Postgres that is configured with a standy streaming replica. 


