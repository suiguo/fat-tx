from sqlalchemy.ext.declarative import *

from sqlalchemy import *

Base = declarative_base()


class Strategy(Base):
    __tablename__ = 't_strategy'
    id = Column(Integer, primary_key=True, autoincrement=True, name='f_id')
    chain = Column(String, name='f_chain')
    project = Column(String, name='f_project')
    currency0 = Column(String, name='f_currency0')
    currency1 = Column(String, name='f_currency1')

    def __str__(self):
        return "chain:{} project:{} currency0:{} currency1:{}".format(self.chain, self.project,
                                                                      self.currency0, self.currency1)


class Currency(Base):
    __tablename__ = 't_currency'
    id = Column(Integer, primary_key=True, autoincrement=True, name='f_id')
    name = Column(String, name='f_name')
    min = Column(Float, name='f_min')
    crossDecimal = Column(Integer, name='f_cross_scale')

    def __str__(self):
        return "name:{} min:{} crossScale:{} tokens:{}".format(self.name, self.min, self.crossDecimal, self.tokens)


class Token(Base):
    __tablename__ = 't_token'
    id = Column(Integer, primary_key=True, autoincrement=True, name='f_id')
    chain = Column(String, name='f_chain')
    currency = Column(String, name='f_currency')
    symbol = Column(String, name='f_symbol')
    address = Column(String, name='f_address')
    decimal = Column(Integer, name='f_decimal')
    crossSymbol = Column(String, name='f_cross_symbol')

    def __str__(self):
        return "chain:{} currency:{} symbol:{} address:{} decimal:{} crossSymbol:{}".format(self.chain, self.currency,
                                                                                            self.symbol, self.address,
                                                                                            self.decimal,
                                                                                            self.crossSymbol)


def find_strategies_by_chain_and_currency(session, chain, currency):
    return session.query(Strategy).filter(Strategy.chain == chain).filter(
        or_(Strategy.currency0 == currency, Strategy.currency1 == currency))


def find_strategies_by_chain_project_and_currencies(session, chain, project, currency0, currency1):
    return session.query(Strategy).filter(Strategy.chain == chain).filter(Strategy.project == project).filter(
        or_(and_(Strategy.currency0 == currency0, Strategy.currency1 == currency1),
            and_(Strategy.currency0 == currency1, Strategy.currency1 == currency0)))


def find_currency_by_address(session, address):
    token = [x for x in session.query(Token).filter(Token.address == address)]
    if len(token) == 0:
        return None

    if len(token) > 1:
        raise ValueError('find more than one token with address:{}', address)

    return token[0].currency