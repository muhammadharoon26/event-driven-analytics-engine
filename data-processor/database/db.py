import os
from sqlalchemy import create_engine, Column, Integer, String, DateTime
from sqlalchemy.orm import declarative_base, sessionmaker

Base = declarative_base()

class AnalyticsEvent(Base):
    __tablename__ = 'analytics_events'

    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(String, index=True, nullable=False)
    action = Column(String, index=True, nullable=False)
    timestamp = Column(DateTime, nullable=False)
    processed_at = Column(DateTime, nullable=False)

def get_engine():
    db_host = os.getenv("DB_HOST", "localhost")
    db_user = os.getenv("DB_USER", "postgres")
    db_password = os.getenv("DB_PASSWORD", "postgres")
    db_name = os.getenv("DB_NAME", "analytics")
    
    # Format for psycopg2
    DATABASE_URL = f"postgresql+psycopg2://{db_user}:{db_password}@{db_host}:5432/{db_name}"
    
    engine = create_engine(DATABASE_URL, echo=False)
    return engine

def init_db():
    engine = get_engine()
    Base.metadata.create_all(bind=engine)
    return engine

def get_session():
    engine = get_engine()
    SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
    return SessionLocal()
