

building docker images
docker build -t api -f cmd/api/Dockerfile  .
docker run -it api /bin/bash


Need to add
[wallet, identity, contract] crud postman testing
create transaction and use it for [issue, redeem and transfer]
nats
custom filter based on fields, search query, ordere by, limit, etc.
standardise reponse using custom err/response using response header and logging
COUCHBASE SDK
select only where isdelte is false
--------- HERE WE ARE ------------
adding autorization in the endpoints
timezone in front end and backend

capture sessioned id
storing time in UTC format as 2023-01-24T16:52:14.90887+00:00
standardise error handling and loggin using a wrapper
nned to put wallet, transaction, idenity in differnet bucket or schema or collection
cashier
jaeger (sp.AddAtrributes)
add swagger documentation


give operator to read only access of the wallet [fronent and backend]
merge wallet [fronent and backend]
mobile version of app and deploy []
staging environment setup on k8s
loyyal node setup
first
channel1 (loyyal -emirates) [wallet, tx, contracts]
channel2 (loyyal -etihad)
channel2 (loyyal -airindia)
the channel0 [loyyal emirates etihad]
