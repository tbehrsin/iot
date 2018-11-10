
cd /tmp
git clone git://github.com/mholling/rpirtscts.git
mv ./rpirtscts/rpirtscts /usr/sbin/rpirtscts

systemctl enable rpirtscts.service
