# ourroots

To build and run this API, you will need a [Go 1.19+ installation](https://golang.org/dl/). 

## Getting Started

Clone the repo:
```
git clone https://github.com/ourrootsorg/cms-server.git
```

## Instructions for running server and client 
#### requires docker-compose, tilt, psql, npm, vue, and a few other things

```
install swag                    # https://github.com/swaggo/swag#getting-started
install migrate                 # https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
install docker                  # https://www.docker.com/ (don't use the snap version)
install docker-compose          # https://docs.docker.com/compose/install/
                                  # no need to install docker-compose on mac since mac docker includes compose
install tilt                    # https://tilt.dev/
                                  # optional but makes rebuilds much faster
                                  # ignore kubernetes and kubectl; you just need the one-line curl install
install npm                     # https://nodejs.org/en/  # v14.21.1 works, v19.2.0 does not
                                  # node includes npm
install psql                    # https://blog.timescale.com/tutorials/how-to-install-psql-on-mac-ubuntu-debian-windows/
npm install -g @vue/cli         # the client uses vue

docker volume create cms_pgdata # do this once to create a persistent database volume
docker volume create cms_s3data # do this once to create a persistent blob store volume
docker volume create cms_esdata # do this once to create a persistent elasticsearch volume

tilt up                         # run the server and dependencies
                                  # make sure you don't already have a postgres process running
                                  # alternatively, run docker-compose up --build
cd db && ./db_setup.sh && ./db_load_core.sh && cd .. # do this once to set up the database
                                  # make sure you have psql (postgres client) available on your path
open http://localhost:19001     # launch the minio browser and create a bucket named "cmsbucket" -- do this once
cd elasticsearch && ./es_setup.sh && cd ..   # do this once to set up elasticsearch
tilt down                       # now that everything is set up, take everything down

tilt up                         # restart everything once it has been set up above
                                  # alternatively, run docker-compose down && docker-compose up --build
cd client                       # the directory for the client
npm install                     # do this once, and again if you get a missing dependency error
npm run serve                   # run the client
                                # make changes to either the server or client, everything reloads automatically
tilt down                       # clean up docker images when done
                                  # alternatively, run docker-compose down
```

## Building 

In the `ourroots` directory, run `make` to run unit tests and build.

TODO: Tests have broken; I'm not sure why; need to investigate. 

## Saving and restoring elasticsearch volume data
### creates /tmp/cms_esdata.tar.bz2 from cms_esdata
```
docker run --rm -v cms_esdata:/volume -v /tmp:/backup alpine tar -cjf /backup/cms_esdata.tar.bz2 -C /volume ./
```
### restores /tmp/cms_esdata.tar.bz2 into cms_esdata
```
docker run --rm -v cms_esdata:/volume -v /tmp:/backup alpine sh -c "rm -rf /volume/* /volume/..?* /volume/.[!.]* ; tar -C /volume/ -xjf /backup/cms_esdata.tar.bz2"
```

## Configuration

### Populating the database tables

To populate the place and name-variants dictionaries, launch a micro instance,
give the micro instance's security group to the database security group,
copy the files in the db drectory to the micro instance, and run the following. You can get the postgres user and password from the secrets manager under postgres/cms/master

First, install the postgres client

```
sudo dnf install postgresql15
```

Next install the migrations tool

```
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
```

Next run migrations
```
./migrate -database "postgres://ourroots_schema:POSTGRES_PASSWORD@POSTGRES_HOST:5432/cms?sslmode=disable" -path migrations up
```

Finally populate the database

```
./db_load_full.sh ourroots_schema POSTGRES_PASSWORD POSTGRES_HOST
```

Don't forget to stop the micro instance when you are finished.

### Populating the ElasticSearch index mapping

First, go to your new Elasticsearch index in AWS, select Security Configuration, edit the Access Policy, and temporarily add the following statement so you can access the ES instance from your local machine. 

```
 {
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "es:ESHttp*",
      "Resource": "ELASTICSEARCH_DOMAIN_ARN/*",
      "Condition": {
        "IpAddress": {
          "aws:SourceIp": "YOUR_LOCAL_IP_GOES_HERE/32"
        }
      }
    }
```

Next, run 

```
curl -X PUT "ELASTICSEARHCH_DOMAIN_ENDPOINT/records" -H 'Content-Type: application/json' -d @elasticsearch/elasticsearch_schema.json
```

Finally, remove what you added to the Elasticsearch Access Policy.

### Customizing the search index

By default, the narrow name coder is NYSIIS, and the broad name coder is Soundex. You can change this by editing 
`elasticsearch/elasticsearch_schema.json` and modifying the encoder values on lines 39 and 43 before running the 
`es_setup.sh` script. Possible values are: nysiis, metaphone, double_metaphone, beider_morse, soundex, refined_soundex, 
daitch_mokotoff, caverphone1, caverphone2, cologne, koelnerphonetik, and haasephonetik. 

## Deploying

First, see "Building" above

See deploy/awslambda/README.md to deploy the server and client, and search-client/README.md to deploy the search client

## Hosting this software on your AWS account

If you want to host this for your society, you need to
* Create an AWS account, and choose an AWS region
* Create a DNS Hosted Zone for the admin domain hosted on AWS Route53
* Create an Auth0 account
  * Create an Application, an API, and set the default audience to your API audience in Settings
* Make the following changes to this software
  * Update OIDC_DOMAIN and OIDC_AUDIENCE in cms-aurora.cf.yaml
    * OIDC_DOMAIN comes from Auth0
    * OIDC_AUDIENCE also comes from Auth0; it can be anything you want; it doesn't even have to exist
* Run `make` to build the software
* Follow the instructions in deploy/awslambda/README.md, replacing app.sbgen-ourroots.com with your admin domain
* Populate the database tables and the elasticsearch index mapping as described above.
