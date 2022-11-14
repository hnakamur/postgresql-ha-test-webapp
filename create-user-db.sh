sudo -iu postgres createuser --pwprompt hatest
# set password to hatestpass

sudo -iu postgres createdb -E utf8 -O hatest hatest
