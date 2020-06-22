# ourroots

To build and run this API, you will need a [Go 1.14+ installation](https://golang.org/dl/). 

## Getting Started

Clone the repo:
```
git clone https://github.com/ourrootsorg/cms-server.git
```

## Instructions for running server and uglyui client 
#### requires docker-compose, tilt, psql, npm, vue, and a few other things

```
install swag                    # https://github.com/swaggo/swag#getting-started
install migrate                 # https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
install docker                  # https://www.docker.com/ 
install docker-compose          # https://docs.docker.com/compose/install/
                                  # no need to install docker-compose on mac since mac docker includes compose
install tilt                    # https://tilt.dev/
                                  # optional but makes rebuilds much faster
                                  # ignore kubernetes and kubectl; you just need the one-line curl install
install npm                     # https://nodejs.org/en/ 
                                  # node includes npm
install psql                    # https://blog.timescale.com/tutorials/how-to-install-psql-on-mac-ubuntu-debian-windows/
npm install -g @vue/cli         # the uglyui client uses vue

docker volume create cms_pgdata # do this once to create a persistent database volume
docker volume create cms_s3data # do this once to create a persistent blob store volume
docker volume create cms_esdata # do this once to create a persistent elasticsearch volume

tilt up                         # run the server and dependencies
                                  # make sure you don't already have a postgres process running
                                  # alternatively, run docker-compose up --build
cd db && ./db_setup.sh && cd .. # do this once to set up the database
                                  # make sure you have psql (postgres client) available on your path
open http://localhost:9000      # launch the minio browser and create a bucket named "cmsbucket" -- do this once
cd elasticsearch && ./es_setup.sh && cd ..   # do this once to set up elasticsearch
tilt down                       # now that everything is set up, take everything down

tilt up                         # restart everything once it has been set up above
                                  # alternatively, run docker-compose down && docker-compose up --build
cd uglyui                       # the directory for the uglyui client
npm install                     # do this once, and again if you get a missing dependency error
npm run serve                   # run the ugly client
                                # make changes to either the server or client, everything reloads automatically
tilt down                       # clean up docker images when done
                                  # alternatively, run docker-compose down
```

# Building 

In the `ourroots` directory, run `make` to run unit tests and build.

# Saving and restoring elasticsearch volume data
### creates /tmp/cms_esdata.tar.bz2 from cms_esdata
```
docker run --rm -v cms_esdata:/volume -v /tmp:/backup alpine tar -cjf /backup/cms_esdata.tar.bz2 -C /volume ./
```
### restores /tmp/cms_esdata.tar.bz2 into cms_esdata
```
docker run --rm -v cms_esdata:/volume -v /tmp:/backup alpine sh -c "rm -rf /volume/* /volume/..?* /volume/.[!.]* ; tar -C /volume/ -xjf /backup/cms_esdata.tar.bz2"
```
