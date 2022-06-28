
# Setting up your development environment

to just start the database container you can run

    docker-compose up

To start a greenlight instance you can combine
the docker files:

    docker-compose -f docker-compose.yml -f greenlight.yml up

As you always have to add both files, using an alias is advised.

## Setting up the database using docker

    # optional
    docker-compose exec -T postgres psql -U postgres -c 'DROP DATABASE b3scale_test'
    docker-compose exec -T postgres psql -U postgres -c 'CREATE DATABASE b3scale'

    # optional
    docker-compose exec -T postgres psql -U postgres -c 'DROP DATABASE b3scale_test'

    docker-compose exec -T postgres psql -U postgres -c 'CREATE DATABASE b3scale_test'

    docker-compose exec -T postgres psql -U postgres -d b3scale < ../db/schema/*.sql           
    docker-compose exec -T postgres psql -U postgres -d b3scale_test < ../db/schema/*.sql           



