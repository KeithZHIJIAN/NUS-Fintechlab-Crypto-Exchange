from decimal import Decimal
import pika


class MQHelper:
    connection = pika.BlockingConnection(pika.ConnectionParameters(host="localhost"))
    channel = connection.channel()

    @classmethod
    def listen_rabbitmq(cls, matchingengine):

        cls.channel.queue_declare(queue="exchange")

        def callback(ch, method, properties, body):
            matchingengine.apply(body)

        cls.channel.basic_consume(
            queue="exchange", on_message_callback=callback, auto_ack=True
        )

        print(" [*] Waiting for messages. To exit press CTRL+C")
        cls.channel.start_consuming()

    @classmethod
    def update_market_history(
        cls,
        time: str,
        symbol: str,
        open: Decimal,
        close: Decimal,
        high: Decimal,
        low: Decimal,
        volume: Decimal,
    ):
        msg = f'{{"time":"{time}","symbol":"{symbol}","open":"{open}","close":"{close}","high":"{high}","low":"{low}","volume":"{volume}"}}'
        print(msg)
        cls.channel.basic_publish(
            exchange="NEW_MARKET_HISTORY.DLQ.Exchange", routing_key="", body=msg
        )

    @classmethod
    def update_ask_order(cls):
        cls.channel.basic_publish(
            exchange="NEW_ASK_ORDER.DLQ.Exchange", routing_key="", body=""
        )

    @classmethod
    def update_bid_order(cls):
        cls.channel.basic_publish(
            exchange="NEW_BID_ORDER.DLQ.Exchange", routing_key="", body=""
        )
