from gql import Client, gql
from gql.transport.requests import RequestsHTTPTransport
import datetime
from requests import post, get
import json

class NUSwapConnector():
    def __init__(self, token):
        self.token = "Bearer " + token 
        self.transport = RequestsHTTPTransport(
                            url="http://localhost:4000/graphql",
                            headers={'Authorization': self.token},
                            verify=True,
                            retries=3,
                        )
        self.client = Client(transport=self.transport, fetch_schema_from_transport=True)
        self.userid = self.getUserDetails()['userid']
        self.walletid = self.getUserDetails()['userid']

# ====================================================================================================
    # Main User Queries
    # Get Details
    def getUserDetails(self):
        query = gql(
                        """
                        query CurrentUser {
                            currentUser {
                                userid
                                name
                                email
                                phone
                                balance
                            }
                        }
                    """
        )

        result = self.client.execute(query)
        return result['currentUser']   
    
    # Get Balance
    def getUserBalance(self):
        query = gql(
                        """
                        query CurrentUser {
                            currentUser {
                                balance
                            }
                        }
                    """
        )

        result = self.client.execute(query)
        return result['currentUser']['balance']  
    
    # Add Balance
    def addUserBalance(self, topUpAmount):
        query = gql(
                        """
                        mutation AddBalance($userid: ID!, $amount: Float!) {
                            addBalance(userid: $userid, amount: $amount) {
                                status
                                response
                                error
                            }
                        }
                    """
        )
        
        variables_dict = {
            'userid': self.userid,
            'amount': topUpAmount
        }
        result = self.client.execute(query, variable_values = variables_dict)
        
        print("New Balance : ", str(self.getUserBalance()))
    
    # Get Wallet Assets
    def getAssets(self):
        query = gql(
                        """
                        query GetWalletAssetsWalletID($walletid: ID!) {
                            getWalletAssetsWalletID(walletid: $walletid) {
                                symbol
                                amount
                            }
                        }
                    """
        )

        variables_dict = {
            'walletid': self.walletid
        }
        result = self.client.execute(query, variable_values = variables_dict)
        
        return result['getWalletAssetsWalletID']
# ====================================================================================================   
    # Order Queries
    # Get Open Ask
    def getOpenAsk(self, symbol):
        query = gql(
                        """
                        query GetOpenAskOrdersForSymbolAndUser($symbol: String!, $owner: ID) {
                            getOpenAskOrdersForSymbolAndUser(symbol: $symbol, owner: $owner) {
                                orderid
                                quantity
                                price
                                openquantity
                                fillcost
                                createdat
                                updatedat
                            }
                        }
                    """
        )

        variables_dict = {
            'symbol': symbol,
            'owner': self.userid
        }
        result = self.client.execute(query, variable_values = variables_dict)['getOpenAskOrdersForSymbolAndUser']
        
        if result is not None:
            for i in result:
                i['createdat'] = datetime.datetime.fromtimestamp(int(i['createdat']) / 1e3).isoformat("T")
                i['updatedat'] = datetime.datetime.fromtimestamp(int(i['updatedat']) / 1e3).isoformat("T")
        
        return result
    
    # Get Open Bid
    def getOpenBid(self, symbol):
        query = gql(
                        """
                        query GetOpenBidOrdersForSymbolAndUser($symbol: String!, $owner: ID) {
                            getOpenBidOrdersForSymbolAndUser(symbol: $symbol, owner: $owner) {
                                orderid
                                quantity
                                price
                                openquantity
                                fillcost
                                createdat
                                updatedat
                            }
                        }
                    """
        )

        variables_dict = {
            'symbol': symbol,
            'owner': self.userid
        }
        result = self.client.execute(query, variable_values = variables_dict)['getOpenBidOrdersForSymbolAndUser']
        
        if result is not None:
            for i in result:
                i['createdat'] = datetime.datetime.fromtimestamp(int(i['createdat']) / 1e3).isoformat("T")
                i['updatedat'] = datetime.datetime.fromtimestamp(int(i['updatedat']) / 1e3).isoformat("T")
            
        return result
    
    # Get Closed
    def getClosed(self, symbol):
        query = gql(
                        """
                        query GetClosedOrdersForSymbolAndUser($symbol: String!, $owner: ID) {
                            getClosedOrdersForSymbolAndUser(symbol: $symbol, owner: $owner) {
                                orderid
                                buyside
                                quantity
                                price
                                fillprice
                                createdat
                                filledat
                            }
                        }
                    """
        )

        variables_dict = {
            'symbol': symbol,
            'owner': self.userid
        }
        result = self.client.execute(query, variable_values = variables_dict)['getClosedOrdersForSymbolAndUser']

        if result is not None:
            for i in result:
                i['createdat'] = datetime.datetime.fromtimestamp(int(i['createdat']) / 1e3).isoformat("T")
                i['filledat'] = datetime.datetime.fromtimestamp(int(i['filledat']) / 1e3).isoformat("T")
            
        return result
    
    # Get All Open
    def getAllOpen(self, symbol):
        openAsks = self.getOpenAsk(symbol)
        openBids = self.getOpenBid(symbol)
        
        result = []
    
        if openAsks is not None:
            for i in openAsks:
                i['buyside'] = 'SELL'
                result.append(i)
                
        if openBids is not None:
            for i in openBids:
                i['buyside'] = 'BUY'
                result.append(i)
            
        return result
    
    # Get All Orders
    def getAllOrders(self, symbol):
        openOrders = self.getAllOpen(symbol)
        closedOrders = self.getClosed(symbol)
        
        result = []
        
        if openOrders is not None:
            for i in openOrders:
                i['status'] = 'OPEN'
                if i['fillcost'] == 0:
                    i['fillprice'] = 0
                else:
                    i['fillprice'] = float(i['fillcost']) / ( float(i['quantity']) - float(i['openquantity']) )
                i['filledat'] = None
                result.append(i)
                
        if closedOrders is not None:
            for i in closedOrders:
                i['status'] = 'CLOSED'
                i['fillcost'] = float(i['fillprice']) * float(i['quantity'])
                i['openquantity'] = 0
                i['updatedat'] = i['filledat']
                result.append(i)
            
        return result
# ====================================================================================================
    # Operations
    # Limit Buy
    # Limit Sell
    # Market Buy
    # Market Sell
    def placeOrder(self, symbol, order_type, side, quantity, price = 0):
        ownerId = self.userid
        walletId = self.walletid
        
        if order_type not in ['Limit','Market']:
            print("ORDER TYPE ERROR. Must be in ",['Limit','Market'])
            return
        if side not in ['Buy','Sell']:
            print("ORDER TYPE ERROR. Must be in ",['Buy','Sell'])
            return
        if quantity <= 0:
            print("quantity must be greater than 0")
            return
        
        # query = gql(
        #                 """
        #                 mutation CreateOrder($symbol: String!, $type: String!, $side: String!, $quantity: Float!, $price: Float!, $ownerId: Int!, $walletId: Int!) {
        #                         createOrder(symbol: $symbol, type: $type, side: $side, quantity: $quantity, price: $price, ownerId: $ownerId, walletId: $walletId) {
        #                             status
        #                             response
        #                             error
        #                         }
        #                     }
        #             """
        # )
        
        # variables_dict = {
        #     'symbol': symbol,
        #     'type': order_type,
        #     'side': side,
        #     'quantity': quantity,
        #     'price': price,
        #     'ownerId': int(ownerId),
        #     'walletId': int(walletId)
        # }
        # result = self.client.execute(query, variable_values = variables_dict)
        
        # symbol := strings.ToUpper(req.Symbol)
        # isBuy := strings.ToUpper(req.Side) == "BUY"
        # quantity := req.Quantity
        # price := req.Price
        # if strings.ToUpper(req.Type) == "MARKET" {
        # 	price = decimal.Zero
        # }
        # ownerId := req.OwnerID
        # walletId := req.WalletID

        result = post("http://localhost:8000/order",json={
            'symbol':symbol, 
            'type': order_type,
            'side': side,
            'quantity': quantity,
            'price': price,
            'owner_id': ownerId,
            'wallet_id': walletId
        })
        return result
    
# ====================================================================================================
    # MarketHistory
    def getMarketHistory(self, symbol, number = 10):
        
        query = gql(
                        """
                        query ReadMarketHistoryAPI($symbol: String!, $number: Int) {
                            readMarketHistoryAPI(symbol: $symbol, number: $number) {
                                time
                                open
                                close
                                high
                                low
                                volume
                            }
                        }
                    """
        )

        variables_dict = {
            'symbol': symbol,
            'number': number
        }
        result = self.client.execute(query, variable_values = variables_dict)['readMarketHistoryAPI']
        
        if result is not None:
            for i in result:
                i['time'] = datetime.datetime.fromtimestamp(int(i['time']) / 1e3).isoformat("T")
                
        return result
    
    # Get Order Fillings
        # Get All order Fillings For Symbol
        
    # Get Open Orders for a symbol
        # Get Open Ask
        # Get Open Bid
    def getAvailableOrders(self, symbol, number = 0):
        
        ask_orders_query = gql(
                            """
                            query GetOpenAskOrdersForSymbol($symbol: String!, $number: Int) {
                                getOpenAskOrdersForSymbol(symbol: $symbol, number: $number) {
                                    price
                                    openquantity
                                    createdat
                                    updatedat
                                }
                            }
                        """
            )

        variables_dict = {
                            'symbol': symbol,
                            'number': number
                        }
        ask_result = self.client.execute(ask_orders_query, variable_values = variables_dict)['getOpenAskOrdersForSymbol']

        if ask_result is not None:
            for i in ask_result:
                i['createdat'] = datetime.datetime.fromtimestamp(int(i['createdat']) / 1e3).isoformat("T")
                i['updatedat'] = datetime.datetime.fromtimestamp(int(i['updatedat']) / 1e3).isoformat("T")
                i['buyside'] = "SELL"

        bid_orders_query = gql(
                            """
                            query GetOpenBidOrdersForSymbol($symbol: String!, $number: Int) {
                                getOpenBidOrdersForSymbol(symbol: $symbol, number: $number) {
                                    price
                                    openquantity
                                    createdat
                                    updatedat
                                }
                            }
                        """
            )

        variables_dict = {
                            'symbol': symbol,
                            'number': number
                        }
        bid_result = self.client.execute(bid_orders_query, variable_values = variables_dict)['getOpenBidOrdersForSymbol']

        if bid_result is not None:
            for i in bid_result:
                i['createdat'] = datetime.datetime.fromtimestamp(int(i['createdat']) / 1e3).isoformat("T")
                i['updatedat'] = datetime.datetime.fromtimestamp(int(i['updatedat']) / 1e3).isoformat("T")
                i['buyside'] = "BUY"
                
        result = []
        if ask_result is not None and bid_result is not None:
            result = ask_result + bid_result
        elif ask_result is not None and bid_result is None:
            result = ask_result
        elif ask_result is None and bid_result is not None:
            result = bid_result
        else:
            result = None
        
        return result
    
    # GetSymbolPrice
    def getPrice(self,symbol):
        return self.getMarketHistory(symbol)[0]['close']


    # Def all Pandas Queries
    




































# ========================== TESTS BELOW COMMENTED ====================
# # Test the stuff

# ## Create

# token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOjEsImVtYWlsIjoiYWxpY2VAZ21haWwuY29tIiwiaWF0IjoxNjU2NDExODA2fQ.vPtq25i87YNCmk4olBZ6ZgmoTEFTfkHIom1jORbB96c"
# NSC = NUSwapConnector(token)

# ## Market History and AvailableOrders

# # NSC.getMarketHistory('btcusd')
# pd.DataFrame(NSC.getMarketHistory('btcusd'))

# NSC.getPrice('btcusd')

# pd.DataFrame(NSC.getAvailableOrders('btcusd'))

# ## User Details

# NSC.getUserDetails()

# NSC.getUserBalance()

# NSC.addUserBalance(100)

# NSC.getAssets()

# ## Order Queries

# # NSC.getOpenAsk('btcusd')
# pd.DataFrame(NSC.getOpenAsk('btcusd'))

# # NSC.getOpenBid('btcusd')
# pd.DataFrame(NSC.getOpenBid('btcusd'))

# # NSC.getClosed('btcusd')
# pd.DataFrame(NSC.getClosed('btcusd'))

# # NSC.getAllOpen('btcusd')
# pd.DataFrame(NSC.getAllOpen('btcusd'))

# # NSC.getAllOrders('btcusd')
# pd.DataFrame(NSC.getAllOrders('btcusd'))

# ## Place 4 Orders

# NSC.placeOrder(symbol = 'btcusd', order_type = 'Market', side = 'Buy', quantity = 1, price = 0)

# NSC.placeOrder(symbol = 'btcusd', order_type = 'Market', side = 'Sell', quantity = 1, price = 0)

# NSC.placeOrder(symbol = 'btcusd', order_type = 'Limit', side = 'Buy', quantity = 1, price = 30)

# NSC.placeOrder(symbol = 'btcusd', order_type = 'Limit', side = 'Sell', quantity = 1, price = 300)

# ## Results after order placing

# pd.DataFrame(NSC.getAllOrders('btcusd'))

# pd.DataFrame(NSC.getMarketHistory('btcusd'))

# pd.DataFrame(NSC.getAvailableOrders('btcusd'))



