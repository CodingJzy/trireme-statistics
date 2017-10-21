# trireme-statistics

trireme-statistics is an aporeto statistics and metrics collector service

## containers launched

  - collector (generate the links and start server)
  - influxdb (stats storage)
  - grafana (metrics visualization)
  - dataexploder (optional, for debugging)

## to start the service

It is easy to start the service and it can be launched using both docker-compose and kubernetes

First, From root directory go into collector
  
    $ cd collector

and build the project using

    $ make install
  
That's it! Now launch the containers using either of the two methods below

## using docker-compose

From root root directory

     $ cd deployments/docker-compose
     
and 

     $ docker-compose up
     
All containers will be launched

Now,

 - goto http://localhost:3000 for grafna dashboards

 - goto http://localhost:8080/graph?address=/get for graph visualization
    

## using kubernetes

From root root directory

     $ cd deployments/kubernetes
and 
    
     $ kubectl create -f .

All containers will be launched as pods in kube cluster

Now,

 - goto http://<externalIP/grafana>:3000 for grafana dashboards

 - goto http://<externalIP/collector>:8080/graph?address=/get for graph visualization
 
