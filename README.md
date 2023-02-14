# Loyyal GinGon
A reworking of the Loyyal ledger based on gin framework that will keep track the information regarding the operator, partner or user identities, their wallets and blockchain transactions. Blockchain ledger is build using the hyperledger fabric version 2.4.7 and it being consumed by the services called `cashier` which receives the events from the `NATS`. Apart from that `api` services is used to expose the endpoints that can be consumed by the partner/operator for intergarion purpose.

- `API`, first most layer where majorlty business logic lies
- `Couchbase`, A NoSQL based database used as the off chain database. For more follow, https://www.couchbase.com/
- `NATS`, is an open-source messaging system (sometimes called message-oriented middleware). The NATS server is written in the Go programming language. For more, follow https://nats.io/
- `Cashier`, a micro-services used to recive the event from the NATS engine and get them publish on to the hyperledger fabric network using fabric sdk v1.0
- `Jaeger`, an open source, end-to-end distributed tracing tool used for monitor and troubleshoot transactions in complex distributed systems For more, follow https://www.jaegertracing.io/
- `Retry`, a micro-service that runs as daemon to settle the transaction which are not pushed to the blockchain ledger



Get the repo:
```
mkdir -p ~/go/src/github.com/loyyal
cd ~/go/src/github.com/loyyal
git clone git@github.com:loyyal-gin-gon/golo.git
cd loyyal-gin-gon
```

### Install dependenies
```
./setup.sh
go get ./...
go get -u golang.org/x/sys
go get -u google.golang.org/grpc


```
### Run the project locally

```
docker-compose -f config/docker-compose.yaml up couch -d
docker-compose -f config/docker-compose.yaml up nats -d
docker-compose -f config/docker-compose.yaml up jaeger1 -d

./cmd/api/run.sh
./cmd/cashier/run.sh

```

### Build the project

```
make build (build the docker images for the golo services)
make push (push the docker images to the erc/dockerhub repository)
```

### To deploy the project

```
make all (delete everything and start golo from strach)
make start (starting the project)
make restart (restrating the project)
make stop (stop the network)
make purge (delete all container, volumes and network)
```

### To create the index in couchbase
```
CREATE PRIMARY INDEX ON `default`:`testbucket`
CREATE PRIMARY INDEX ON `default`:`testbucket`.`_default`.`users`
```

#
## Environment Listing

| Environment | App Url | APi Url |
| --- | --- | --- |
| Production | https://dashboard.loyyal.net | https://api.loyyal.net |
| Staging | https://dashboard.staging.loyyal.net | https://api.staging.loyyal.net |
| Development | https://dashboard.dev.loyyal.net | https://api.dev.loyyal.net |


Postman API Documentation (https://documenter.getpostman.com/view/24557333/2s8YzRzi69)


## References
- https://go.dev/
- https://hyperledger-fabric.readthedocs.io/en/release-2.5/
- https://docs.couchbase.com/home/index.html
- https://docs.couchbase.com/server/current/learn/data/scopes-and-collections.html
- https://stackoverflow.com/questions/71507321/go-1-18-build-error-on-mac-unix-syscall-darwin-1-13-go253-golinkname-mus
- 

# Copyright and license

The loyyal suite is Copyright 2020-2023.

Licensed under the [Apache License, Version 2.0][license] (the "License");
you may not use this software except in compliance with the License.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

[license]: http://www.apache.org/licenses/LICENSE-2.0