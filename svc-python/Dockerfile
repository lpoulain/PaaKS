FROM python:3.8-slim-buster
 
# Creating Application Source Code Directory
#RUN mkdir -p /svc/src
# Setting Home Directory for containers
WORKDIR /usr/src/app
# Installing python dependencies
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
# Copying src code to Container
COPY app.py app.py
# Application Environment variables
#ENV APP_ENV development
# Exposing Ports
EXPOSE 5000
RUN ln -s /tmp/storage/ /usr/src/app/lib
# Setting Persistent data
#VOLUME [“/app-data”]
# Running Python Application
CMD python app.py
