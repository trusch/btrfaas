FROM alpine

ADD https://github.com/trusch/btrfaas/releases/download/v0.2.1/frunner /bin/frunner
RUN chmod +x /bin/frunner

ADD script.sh /script.sh
RUN chmod +x /script.sh

CMD ["/bin/frunner","--","/script.sh"]
