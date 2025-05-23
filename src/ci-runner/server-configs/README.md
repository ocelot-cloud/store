### Ocelot Server Setup

Simple and secure setup to run services with an unprivileged user.

* Create user with minimal rights:

```bash
ssh server # as root
useradd --no-create-home --shell /usr/sbin/nologin --user-group user
sudo -u user echo lol # test
```

* add crontab configs and `/usr/local/sbin/update.sh`
* configure ufw:

```bash
ufw default deny incoming
ufw default allow outgoing
ufw allow 22
ufw allow 80
ufw allow 443
ufw enable
```

* copy docker-compose.yml and dynamic.yml to server:/root and run `docker-compose up -d`
* add `/etc/systemd/system/store.service` and run:

```bash
systemctl enable --now store
```

* then enter the host and email account data in `store/data/.env` - note that you have to use port 587, because Hetzner blocks port 465 by default, see [here](https://www.reddit.com/r/hetzner/comments/16i4ucp/initial_blocking_of_sending_emails_is_this/?tl=de)
* restart store:

```bash
systemctl restart store
```