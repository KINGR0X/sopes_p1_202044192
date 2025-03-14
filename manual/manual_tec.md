# Manual tecnico

## Requisitos del sistema

- Sistemas Linux
- Minimo 4 gb de RAM
- Procesador de minimo 2 nucleos
- Docker
- Python
- Rust

## Descripcion general

El objetivo de este proyecto es aplicar todos los conocimientos adquiridos en la unidad 1, con la
implementación de un gestor de contenedores mediante el uso de scripts, módulos de kernel,
lenguajes de programación y la herramienta para la creación y manejo de contenedores más popular,
Docker. Con la ayuda de este gestor de contenedores se podrá observar de manera más detallada los
recursos y la representación de los contenedores a nivel de procesos de Linux y como de manera
flexible pueden ser creados, destruidos y conectados por otros servicios.

## Modulo de kernel

El modulo de kernel esta escrito en el lenguaje de c. Dicho modulo es el encargado de capturar las metricas del sistema tales como:

- Total de memoria RAM en mb
- Memoria RAM libre
- Memoria RAM en uso
- Uso del CPU

Tambien obtiene la siguiente información de los contenedores de docker:

- PID
- Nombre
- ID del contenedor
- Porcentaje de memoria RAM utilizada
- Porcentaje CPU utilizado
- Uso del disco de lectura y escritura en mb
- Información de lectura y escritura de i/o

A continuación se muestra un extracto del codigo del modulo de kernel que se encuentra en el archivo inf.c:

```c

static char *extract_container_id(const char *cmdline) {
    char *id_start = strstr(cmdline, "-id ");
    if (!id_start) return NULL; // Si no se encuentra "-id", retornar NULL

    id_start += 4; // Moverse al inicio del ID (después de "-id ")
    char *id_end = strchr(id_start, ' '); // Buscar el siguiente espacio después del ID
    if (!id_end) return NULL; // Si no hay espacio, el ID llega hasta el final

    // Copiar el ID en un nuevo buffer
    int id_length = id_end - id_start;
    char *container_id = kmalloc(id_length + 1, GFP_KERNEL);
    if (!container_id) return NULL;

    strncpy(container_id, id_start, id_length);
    container_id[id_length] = '\0'; // Asegurar que esté terminado en null

    return container_id;
}

static char *get_process_cmdline(struct task_struct *task) {
    struct mm_struct *mm;
    char *cmdline, *p;
    unsigned long arg_start, arg_end, env_start;
    int i, len;

    cmdline = kmalloc(MAX_CMDLINE_LENGTH, GFP_KERNEL);
    if (!cmdline) return NULL;

    mm = get_task_mm(task);
    if (!mm) {
        kfree(cmdline);
        return NULL;
    }

    down_read(&mm->mmap_lock);
    arg_start = mm->arg_start;
    arg_end = mm->arg_end;
    env_start = mm->env_start;
    up_read(&mm->mmap_lock);

    len = arg_end - arg_start;
    if (len > MAX_CMDLINE_LENGTH - 1) len = MAX_CMDLINE_LENGTH - 1;

    if (access_process_vm(task, arg_start, cmdline, len, 0) != len) {
        mmput(mm);
        kfree(cmdline);
        return NULL;
    }

    cmdline[len] = '\0';
    p = cmdline;

    for (i = 0; i < len; i++) {
        if (p[i] == '\0') p[i] = ' ';
    }

    mmput(mm);
    return cmdline;
}
```

## Cronjob

El cronjob se ejecuta cada minuto, y es el encargado de crear 10 contenedores aleatorios, que estresan el sistema autilizando la imagen de alpine-stress. Para el cronjob utilizo dos archivos de bash, los cuales son **crear_cronjob.sh** y **cronjob.sh**

como su nombre indica crear cronjob ejecuta el archivo cronjob.sh cada minuto,e incluye una verificacion para evitar tener más de un cronjob a la vez, ya que podria sobrecargar el sistema.

```sh
#!/bin/bash

# Ejecutar cada minuto

cron_command="*/1 * * * * cd /home/elian/Descargas/sopes_p1_202044192/scrips; ./cronJob.sh"

# Verifica si el cron job ya existe para evitar duplicados
(crontab -l | grep -F "$cron_command") || (crontab -l; echo "$cron_command") | crontab -

```

El archivo cronjob.sh crea 10 contenedores con la imagen de alpine-stress, como argumentos escoge uno de la lista de argumentos de commands, y genera 10 contenedores con dichos argumentos, estas argumentso sirven para configurar si el contenedor estresara la CPU, RAM, I/O, o el disco duro.

```sh
#!/bin/bash

# Definir los comandos como listas de argumentos
commands=(
    "stress --cpu 1 --timeout 600"  # Consume CPU
    "stress --vm 1 --vm-bytes 512M --timeout 600"  # Consume RAM
    "stress --io 1 --timeout 600"  # Consume I/O
    "stress --hdd 1 --hdd-bytes 1G --timeout 600"  # Consume Disco
)

# Generar 10 contenedores de manera aleatoria
for i in {1..10}
do
    # Seleccionar un comando aleatorio
    command=${commands[$RANDOM % ${#commands[@]}]}

    # Generar un nombre único para el contenedor usando /dev/urandom
    container_name="container_$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 8 | head -n 1)"

    # Ejecutar el contenedor con el comando seleccionado
    docker run -d --name $container_name containerstack/alpine-stress $command
done

```

## Servicio de Rust

El servicio de Rust es el corazón del proyecto, este es el encargado de crear al inicio un contenedor que maneja las peticiones https para poder guardar los logs, dichos logs son utilizados para las graficas que se generan al finalizar el programa. ademas de ello es el encargado de eliminar los contendos que genera el cronjob, dejando siempre en ejecución al contenedor que ejecuta los logs, al contedor más reciente y a un contenedor de cada tipo (CPU, RAM, I/O, disco)

A continuación un extracto del codigo de Rust, que se encuentra en el archivo main.rs

```rust
#[derive(Debug, Serialize, Deserialize)]
struct SystemInfo {
    ram_total_mb: f64,
    ram_free_mb: f64,
    ram_usage_mb: f64,
    cpu_usage: f64,
    #[serde(rename = "processes")]
    processes: Vec<Process>,
}

#[derive(Debug, Serialize, Deserialize, PartialEq)]
struct Process {
    #[serde(rename = "PID")]
    pid: u32,
    #[serde(rename = "Name")]
    name: String,
    #[serde(rename = "ContainerID")]
    container_id: String,
    #[serde(rename = "MemoryUsage")]
    memory_usage: f64,
    #[serde(rename = "CPUUsage")]
    cpu_usage: f64,
    #[serde(rename = "DiskUsageRead_mb")]
    disk_usage_read_mb: f64,
    #[serde(rename = "DiskUsageWrite_mb")]
    disk_usage_write_mb: f64,
    #[serde(rename = "IORead_mb")]
    io_read_mb: f64,
    #[serde(rename = "IOWrite_mb")]
    io_write_mb: f64,
}

#[derive(Debug, Serialize, Clone)]
struct LogProcess {
    pid: u32,
    name: String,
    container_id: String,
    memory_usage: f64,
    cpu_usage: f64,
    disk_usage_read_mb: f64,
    disk_usage_write_mb: f64,
    io_read_mb: f64,
    io_write_mb: f64,
    action: String,
    timestamp: String,
    creation_time: Option<String>,
    deletion_time: Option<String>
}
```

Como se menciono con anterioridad el programa de Rust genera al inicio un contendor que procesa los Logs y las peticiones HTTP, para ello se hace uso de docker compose, y python con fastapi.

**docker-compose.yaml**

```yaml
services:
python_service:
build: ./
container_name: logs_container
ports: - 8000:8000
volumes: - ./logs:/code/logs - ./graphs:/code/graphs - .:/code

    command: ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]

```

**DockerFile**

```DockerFile
FROM python:3.12-slim

WORKDIR /code

# Instalar dependencias del sistema para matplotlib
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

COPY ./requirements.txt /code/requirements.txt
RUN pip install --no-cache-dir --upgrade -r /code/requirements.txt

COPY . /code/

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]

```

El archivo main.py se carga al contenedor y se utilizad para generar la API, lo mismo con models.py

```python
@app.get("/graph")
def get_graphs():
    ruta_memory = "./logs/memory.json"
    ruta_cont = "./logs/logs.json"

    try:
        with open(ruta_memory, 'r', encoding='utf-8') as file_memory:
            datos_memory = json.load(file_memory)

        with open(ruta_cont, 'r', encoding='utf-8') as file_cont:
            datos_cont = json.load(file_cont)

    except Exception as e:
        print(f"Error reading files: {e}")
        return {"error": "Failed to read log files"}

    # Grafica de uso de RAM
    x_memory = range(len(datos_memory))

    y_memory = [(elem.get('usage_ram_mb', 0) / elem.get('total_ram_mb', 1))
                * 100 for elem in datos_memory]
```

## Instalación y ejecución del programa

### Modulo de kernel

El modulo de kernel ya esta compilado, pero en caso de que se quiera volver a compilar ya sea por compatibilidad o por otra razón se puede compilar de la siguiente manera (se debe de arbir una terminal en la direccion donde esta el archivo inf.c)

```bash
make clean
make
sudo insmod inf.ko
cat /proc/sysinfo_202044192
```

Con estos comandos se limpiaran los archivos ya compilados, se compilara el modulo y se agregara el modulo a los modulos de Linux. con Cat se puede ver la información que captura el modulo

## Ejecutar el servicio de Rust

Para evitar el riesgo de saturar el sistema de contenedores, primero se debe de ejecutar el sistema de gestión de contenedores de Rust,

Se debe de abrir una terminal en la carpeta donde se encuentra el archivo main.rs (/home/elian/Descargas/sopes_p1_202044192/servicio_rust/src), luego se debe de ejecutar el siguiente comando

```bash
cargo run
```

## Crear el Cronjob

Por ultimo creamos el cronjob, para ello se debe de abrir un la carpeta donde se encuentra el archivo de crear_cronjob.sh (/home/elian/Descargas/sopes_p1_202044192/scrips), luego se ejecuta el siguiente comando:

```bash
sudo ./crearCronjob.sh
```

Para comprobar que si se creo el cronjob podemos abrir una nueva terminal y usar el siguiente comando:

```bash
crontab -l
```

Si el contab se ceo se mirara la siguiente linea al final de todo:

```bash
*/1 * * * * cd /home/elian/Descargas/sopes_p1_202044192/scrips; ./cronJob.sh
```
