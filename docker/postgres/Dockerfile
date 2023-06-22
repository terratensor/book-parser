FROM postgres:15.1-alpine

WORKDIR /app

RUN whoami

COPY sql/init_book-parser.sql /docker-entrypoint-initdb.d/

ENTRYPOINT ["docker-entrypoint.sh"]
EXPOSE 5432
CMD ["postgres"]
