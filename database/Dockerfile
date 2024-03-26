FROM postgres

ENV POSTGRES_PASSWORD postgres
ENV POSTGRES_DB test

COPY schem.sql /docker-entrypoint-initdb.d/
