from src.orderbook import OrderBook

from decimal import Decimal

class MatchingEngine(object):
    __instance = None

    @classmethod
    def get_instance(cls):
        """creates a singleton instance of MatchingEngine

        Args: None
        Returns:
            MatchingEngine instance
        """
        if MatchingEngine.__instance is None:
            MatchingEngine()
        return MatchingEngine.__instance

    def __init__(self):
        """inits MatchingEngine

        Args: None
        Returns: None
        Raises:
            Exception -> This is a singleton.
        """
        if MatchingEngine.__instance is not None:
            raise Exception("This is a singleton.")

        self._orderBooks = dict()
        MatchingEngine.__instance = self

    def apply(self, msg):
        """
        Parse a message into a dictionary of order, and apply it to the order book.
        Args:
            msg (str): Message to parse.
        Returns:
            None
        """
        print(" [x] Received %r" % msg)
        msg_list = msg.decode("UTF-8").split()
        if msg_list[0].upper() == "ADD":
            self.doAdd(msg_list)
        # if msg_list[0] == "CANCEL":
        #     self.doCancel(msg_list)
        # if msg_list[0] == "MODIFY":
        #     self.doModify(msg_list)


        # orderId: str,
        # symbol: str,
        # orderType: str,
        # buy: bool,
        # quantity: Decimal,
        # price: Decimal,
        # stopPrice: Decimal,
        # ownerId: str,
        # walletId: str,
        # creationTime: str,
        # lastModTime: str,  # timestamp of order last modification

        # Modify, Symbol, Order ID, Quantity, Price
        if msg_list[0].upper() == "MODIFY":
            self.doModify(msg_list)
        if msg_list[0].upper() == "CANCEL":
            self.doCancel(msg_list)

    def doAdd(self, msg_list):
        """
        Parse a message into a dictionary of order, and apply it to the order book.
        Args:
            msg (str): Message to parse.
        Returns:
            None
        """
        symbol = msg_list[1].upper()

        if symbol not in self._orderBooks:
            self._orderBooks[symbol] = OrderBook(symbol)
        orderbook = self._orderBooks[symbol]

        orderType = msg_list[2]

        if msg_list[3].upper() == "BID" or msg_list[3].upper() == "BUY":
            buy = True
        elif msg_list[3].upper() == "ASK" or msg_list[3].upper() == "SELL":
            buy = False
        else:
            print("--Invalid side.")
            # Send error msg back to front end
            return False

        quantity = round(Decimal(msg_list[4]),3)
        if quantity == 0 or quantity > 1000000000:
            print("--Invalid quantity")
            return False

        price = Decimal(0) if orderType.upper() == "MARKET" else round(Decimal(msg_list[5]),3)
        if price > 1000000000:
            print("--Invalid price")
            return False

        ownerId = msg_list[6]

        walletId = msg_list[7]

        stopPrice = round(Decimal(msg_list[8]),3) if len(msg_list) > 8 else Decimal(0)

        orderbook.add(symbol, orderType, buy, quantity, price, stopPrice, ownerId, walletId)

        print(orderbook)

    def doModify(self, msg_list):
        """
        -   Modify, Symbol, Side, Order ID, prev Quantity, prev Price, new Quantity, new Price
            modify ETHUSD buy 0000000002 100 63000 100 64000 //change price to 64000 only
        """

        symbol = msg_list[1].upper()

        if symbol not in self._orderBooks:
            print("--Invalid symbol")
            return
        orderbook = self._orderBooks[symbol]

        isBuy = msg_list[2].upper() == "BID" or msg_list[2].upper() == "BUY"

        orderId = msg_list[3]


        prevQuantity = round(Decimal(msg_list[4]),3)
        prevPrice = round(Decimal(msg_list[5]),3)
        newQuantity = round(Decimal(msg_list[6]),3)
        newPrice = round(Decimal(msg_list[7]),3)

        orderbook.replace(orderId, isBuy, prevQuantity, prevPrice, newQuantity, newPrice)

        print(orderbook)

    def doCancel(self, msg_list):
        """
        -   Cancel, Symbol, Side, Order ID, Price
            cancel ETHUSD buy 0000000001 100
        """
        symbol = msg_list[1].upper()

        if symbol not in self._orderBooks:
            print("--Invalid symbol")
            return

        orderbook = self._orderBooks[symbol]

        isBuy = msg_list[2].upper() == "BID" or msg_list[2].upper() == "BUY"

        orderId = msg_list[3]

        price = round(Decimal(msg_list[4]),3)

        orderbook.cancel(isBuy, price, orderId)
        print(orderbook)

    #     quantity = int(msg_list[4])

    #     if quantity > 1000000000:
    #         print("--Invalid quantity")
    #         return False

    #     price = Decimal(msg_list[5])
    #     if price > 1000000000:
    #         print("--Invalid price")
    #         return
    #     orderbook.replace(
    #         order, quantity, price
    #     )

    #     print(orderbook)
