# nce-matchingengine

### Setup Process
* ```python3 init.database.py```
* ```python3 main.py```
* ```python3 init.orderbook.py```

### RabbitMQ Setup 
* ```docker run --rm -it --name mq -d -p 5672:5672 -p 15672:15672 -dit rabbitmq```
