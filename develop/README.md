
# Setting up your development environment

to just start the database container you can run

    docker-compose up

To start a greenlight instance you can combine
the docker files:

    docker-compose -f docker-compose.yml -f greenlight.yml up

As you allways have to add both files, using an alias is advised.
