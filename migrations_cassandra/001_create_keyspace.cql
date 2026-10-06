CREATE KEYSPACE IF NOT EXISTS geoguide
WITH REPLICATION = {
    'class' : 'SimpleStrategy',
    'replication_factor' : 1
};
