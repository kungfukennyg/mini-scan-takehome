FROM postgres:latest

# Set environment variables
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres
ENV POSTGRES_DB=scanning

# Setup initial database
COPY postgres/initdb.sql /docker-entrypoint-initdb.d/

# Expose the PostgreSQL port
EXPOSE 5432

# Set the default command
CMD ["postgres"]