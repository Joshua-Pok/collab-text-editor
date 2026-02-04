<!--markdownlint-disable-->


# Pub Sub

Publish - Subscribe is a architectural pattern is used where many applications talk to each other.


In a point to point, a application sends a message to another

In a pub sub, a application sends a message to a broker, the broker is then responsible for sending that message to subscribers 

Think email vs instagram post


Pub Sub systems are often used to enable event driven design or event driven architecture. Event driven architecture uses events to trigger and communicate between decoupled systems


RabbitMQ

RabbitMQ is a popular open source message broker that implements the AMQP protoccol. 

RabbitMQ has 3 components:

1) Clieint
2) RabbitMQ Server
3) 


## RabbitMQ vs MQTT.

Rabbit MQ is a software that menages message queues, mqtt is a protocol for messaging

RabbitMQ supports AMQP where mosquitto uses MQTT
