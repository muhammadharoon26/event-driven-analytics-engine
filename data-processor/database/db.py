import os
from sqlalchemy import create_engine, Column, Integer, String, DateTime
from sqlalchemy.orm import declarative_base, sessionmaker

Base = declarative_base()

# The engine owns the connection pool, so it must be built once and reused for
# the life of the process. Building one per message means a fresh TCP connect
# and Postgres auth handshake for every event, which caps the consumer at a few
# tens of events per second.
_engine = None
_SessionLocal = None

class AnalyticsEvent(Base):
    __tablename__ = 'analytics_events'

    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(String, index=True, nullable=False)
    action = Column(String, index=True, nullable=False)
    timestamp = Column(DateTime, nullable=False)
    processed_at = Column(DateTime, nullable=False)

def get_engine():
    """Return the process-wide engine, building it on first use."""
    global _engine
    if _engine is not None:
        return _engine

    db_host = os.getenv("DB_HOST", "localhost")
    # Inside Docker this is 5432 (container port); running the worker natively
    # against docker-compose it is 5433, the host port Postgres is published on.
    db_port = os.getenv("DB_PORT", "5432")
    db_user = os.getenv("DB_USER", "postgres")
    db_password = os.getenv("DB_PASSWORD", "postgres")
    db_name = os.getenv("DB_NAME", "analytics")

    # Format for psycopg2
    DATABASE_URL = f"postgresql+psycopg2://{db_user}:{db_password}@{db_host}:{db_port}/{db_name}"
    
    _engine = create_engine(
        DATABASE_URL,
        echo=False,
        # Keep a small warm pool instead of reconnecting per event.
        pool_size=5,
        max_overflow=10,
        # Drop connections the database silently closed, so a long-idle worker
        # does not raise on its next write.
        pool_pre_ping=True,
        pool_recycle=1800,
    )
    return _engine

def init_db():
    engine = get_engine()
    Base.metadata.create_all(bind=engine)
    return engine

def get_session():
    """Check out a session bound to the shared pool."""
    global _SessionLocal
    if _SessionLocal is None:
        _SessionLocal = sessionmaker(
            autocommit=False, autoflush=False, bind=get_engine()
        )
    return _SessionLocal()
