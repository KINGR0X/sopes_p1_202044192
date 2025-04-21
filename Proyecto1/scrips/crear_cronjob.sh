#!/bin/bash

# Ejecutar cada minuto

cron_command="*/1 * * * * cd /home/elian/Descargas/sopes_p1_202044192/scrips; ./cronJob.sh"

# Verifica si el cron job ya existe para evitar duplicados
(crontab -l | grep -F "$cron_command") || (crontab -l; echo "$cron_command") | crontab -