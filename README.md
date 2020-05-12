# ourroots

To build and run this API, you will need a [Go 1.14+ installation](https://golang.org/dl/). 

## Prerequisites

* [Swag](https://github.com/swaggo/swag#getting-started)

To use with a real database:
* [Postgres](https://www.postgresql.org/download/)
* [Migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

## Building and Running

Clone the repo:
```
git clone https://github.com/ourrootsorg/cms-server.git
```
In the `ourroots` directory, run `make` to run unit tests and build.

To run the server using the in-memory 'database':
```
cd server/
PERSISTER=memory ./server
```

You should be able to access the server at http://localhost:3000/. 

## Database Setup

After installing Postgres, cd into the `db` directory and run `./db_setup.sh` which should create the `cms` database and apply database migrations to create tables. Once that is done, you should be able to run the server using the database:
```
PERSISTER=sql DATABASE_URL=postgres://ourroots:password@localhost:5432/cms?sslmode=disable ./server
```

## Instructions for running server and uglyui client 
#### requires docker-compose, tilt, npm, and vue

```
install docker                  # https://www.docker.com/ - includes docker-compose on Mac
install tilt                    # https://tilt.dev/ - makes rebuilds much faster
install npm                     # https://nodejs.org/en/ - includes npm
npm install -g @vue/cli         # the uglyui client uses vue

docker volume create cms_pgdata # do this once to create a persistent database volume
tilt up                         # run the server and dependencies
cd db && ./db_setup.sh          # do this once to set up the database
cd ../uglyui                    # the directory for the uglyui client
npm install                     # do this once, and again if you get a missing dependency error
npm run serve                   # run the ugly client
                                # make changes to either the server or client, everything reloads automatically
tilt down                       # clean up docker images when done
```