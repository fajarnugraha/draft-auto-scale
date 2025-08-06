# Use the official Nginx image from Docker Hub.
FROM nginx:1.21-alpine

# Remove the default Nginx configuration file.
RUN rm /etc/nginx/conf.d/default.conf

# Copy our custom load balancer configuration into the container.
COPY nginx.conf /etc/nginx/conf.d/
