#!/bin/bash

cron_command="*/1 * * * * cd /home/elian/Descargas/sopes_p1_202044192/scrips; ./cronJob.sh"

# elimina cron_command
(crontab -l | grep -v -F "$cron_command") | crontab -
