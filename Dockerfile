FROM nginx:1.22
COPY nginx/default.conf /etc/nginx/conf.d/default.conf
COPY static/json-to-go /usr/share/nginx/html/json-to-go