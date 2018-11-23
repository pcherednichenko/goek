# Go Elasticsearch Kibana example

![logoGOEK](./.github/logoGOEK.png)

In this repo you can find example how to work with Elasticsearch and Go together

### Using technologies

- Golang 1.11
- Go Mod for Golang vendor
- [olivere/elastic.v5](https://github.com/olivere/elastic) library
- Elasticsearch v6.5.1
- Kibana v6.5.1
- Docker and docker-compose for running

### Service does

1. Run ElasticSearch
2. Run Kibana
3. Configurate Kibana dashboards
4. Go application create index in Elasticsearch for test data
5. Go application send data to Elasticsearch with random numbers by random users
6. Go application select data by user name Pavel (search example) and print it in console logs

### How to run

1. Clone project: 
```
go get github.com/pcherednichenko/goek
```

2. Open to project folder:
```
cd $GOPATH/src/github.com/pcherednichenko/goek/
```

3. Run services:
```
docker-compose up -d
```

4. Open http://localhost:5601/app/kibana#/dashboard/065f0890-ef00-11e8-b390-110dd8eed153
to see dashboard data

You can also see result of go applications with command:
```
docker-compose logs -f golang
```