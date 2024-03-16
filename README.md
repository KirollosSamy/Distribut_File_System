# Distribut_File_System
A simple implementation of a distributed file system using gRPC and Golang

- To build the images use

`docker-compose build`

- To start all the containers run

`docker-compose up -d`

Note: this will run the containers in the background. To get inside a specific cotainer you can use

`docker exec -it {container_name} bash`

- To start only the master

`docker-compose up master`

- To start only the client

`docker-compose up client`

- To start node n

`docker-compose up node{n}`

- To stop all the containers run

`docker-compose down`
