use std::fs::File;
use std::io::{self, Read};
use std::path::Path;
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Local};
use reqwest::Client;
use std::error::Error;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::collections::HashMap;
use chrono::TimeZone;
use ctrlc;

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
}

#[derive(Debug, Serialize, Clone)]
struct LogRam {
    total_ram_mb: f64,
    free_ram_mb: f64,
    usage_ram_mb: f64,
    timestamp: String,
}

fn get_container_command(container_id: &str) -> Result<String, Box<dyn Error>> {
    let output = std::process::Command::new("docker")
        .arg("inspect")
        .arg("--format={{.Config.Cmd}}")
        .arg(container_id)
        .output()?;

    if !output.status.success() {
        return Err(format!("Failed to inspect container {}", container_id).into());
    }

    Ok(String::from_utf8(output.stdout)?.trim().to_string())
}

fn get_container_creation_time(container_id: &str) -> Result<DateTime<Local>, Box<dyn Error>> {
    let output = std::process::Command::new("docker")
        .arg("inspect")
        .arg("--format={{.Created}}")
        .arg(container_id)
        .output()?;

    if !output.status.success() {
        return Err(format!("Failed to get creation time for {}", container_id).into());
    }

    let time_str = String::from_utf8(output.stdout)?.trim().to_string();
    DateTime::parse_from_rfc3339(&time_str)
        .map(|dt| dt.with_timezone(&Local))
        .map_err(|e| e.into())
}

fn get_container_name(container_id: &str) -> Result<String, Box<dyn Error>> {
    let output = std::process::Command::new("docker")
        .arg("inspect")
        .arg("--format={{.Name}}")
        .arg(container_id)
        .output()?;

    if !output.status.success() {
        return Err(format!("Failed to get name for {}", container_id).into());
    }

    Ok(String::from_utf8(output.stdout)?
        .trim()
        .trim_start_matches('/')
        .to_string())
}

impl Process {
    fn get_container_id(&self) -> &str {
        &self.container_id
    }
}

impl Eq for Process {}

impl Ord for Process {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        self.cpu_usage.partial_cmp(&other.cpu_usage).unwrap_or(std::cmp::Ordering::Equal)
            .then_with(|| self.memory_usage.partial_cmp(&other.memory_usage).unwrap_or(std::cmp::Ordering::Equal))
    }
}

impl PartialOrd for Process {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        Some(self.cmp(other))
    }
}

fn kill_container(process: &Process) {
    let container_id = process.container_id.clone();
    
    // Obtener comando y tipo
    let command = get_container_command(&container_id).unwrap_or_default();
    let container_type = if command.contains("--cpu") {
        "CPU"
    } else if command.contains("--vm") {
        "Memory"
    } else if command.contains("--io") {
        "I/O"
    } else if command.contains("--hdd") {
        "Disk"
    } else {
        "Unknown"
    };

    println!("Killing container: ID={}, Type={}", container_id, container_type);

    let output = std::process::Command::new("sudo")
        .arg("docker")
        .arg("rm")
        .arg("-f")
        .arg(&container_id)
        .output()
        .expect("failed to execute process");

    if !output.status.success() {
        eprintln!("Error killing container: {}", String::from_utf8_lossy(&output.stderr));
    }
}

fn read_proc_file(file_name: &str) -> io::Result<String> {
    let path = Path::new("/proc").join(file_name);
    let mut file = File::open(path)?;
    let mut content = String::new();
    file.read_to_string(&mut content)?;
    Ok(content)
}

fn parse_proc_to_struct(json_str: &str) -> Result<SystemInfo, serde_json::Error> {
    let system_info: SystemInfo = serde_json::from_str(json_str)?;
    Ok(system_info)
}

#[tokio::main]
async fn enviar_json_logs(logs_procesess: &Vec<LogProcess>) -> Result<(), Box<dyn Error>> {
    let client = Client::new();
    let response = client.post("http://0.0.0.0:8000/logs")
        .json(logs_procesess)
        .send()
        .await?;

    if response.status().is_success() {
        println!("¡Datos enviados exitosamente!");
    } else {
        println!("Error al enviar los datos: {:?}", response.status());
    }

    Ok(())
}

#[tokio::main]
async fn enviar_json_ram(log_ram: &Vec<LogRam>) -> Result<(), Box<dyn Error>> {
    let client = Client::new();
    let response = client.post("http://0.0.0.0:8000/memory")
        .json(log_ram)
        .send()
        .await?;

    if response.status().is_success() {
        println!("¡Datos enviados exitosamente!");
    } else {
        println!("Error al enviar los datos: {:?}", response.status());
    }

    Ok(())
}

#[tokio::main]
async fn get_graph() -> Result<(), Box<dyn Error>> {
    let body = reqwest::get("http://0.0.0.0:8000/graph")
        .await?
        .text()
        .await?;

    println!("body = {body:?}");
    Ok(())
}

fn print_memory_info(system_info: &SystemInfo) {
    println!("==========================================================================================================");
    println!("Memory Information:");
    println!("  Total RAM: {:.2} MB", system_info.ram_total_mb);
    println!("  Free RAM: {:.2} MB", system_info.ram_free_mb);
    println!("  RAM Usage: {:.2} MB", system_info.ram_usage_mb);
    println!();
}

fn print_containers_by_category(containers_by_type: &HashMap<String, Vec<(Process, DateTime<Local>)>>) {
    println!("Containers por categoria:");
    for (category, containers) in containers_by_type {
        println!("  {}:", category);
        for (process, creation_time) in containers {
            println!("    ID: {}, Name: {}, Created: {}", process.container_id, process.name, creation_time);
        }
    }
    println!();
}

fn analyzer(system_info: SystemInfo, logs_container_id: &str) {
    let mut log_proc_list = Vec::new();
    let mut log_ram_list = Vec::new();
    let mut containers_by_type: HashMap<String, Vec<(Process, DateTime<Local>)>> = HashMap::new();

    // Imprimir información de la memoria
    print_memory_info(&system_info);

    for process in system_info.processes {
        let container_id = process.container_id.clone();
        
        // Saltar contenedor de logs
        if container_id == logs_container_id {
            continue;
        }

        // Obtener metadata del contenedor
        let Ok(command) = get_container_command(&container_id) else { continue };
        let Ok(creation_time) = get_container_creation_time(&container_id) else { continue };
        let Ok(container_name) = get_container_name(&container_id) else { continue };

        // Determinar tipo de contenedor
        let container_type = if command.contains("--cpu") {
            "cpu"
        } else if command.contains("--vm") {
            "memory"
        } else if command.contains("--io") {
            "io"
        } else if command.contains("--hdd") {
            "disk"
        } else {
            continue;
        };

        // Agrupar por tipo
        containers_by_type
            .entry(container_type.to_string())
            .or_default()
            .push((process, creation_time));
    }

    // Imprimir contenedores agrupados por categoría
    print_containers_by_category(&containers_by_type);

    // Procesar cada tipo de contenedor
    for (container_type, mut containers) in containers_by_type {
        // Ordenar por fecha de creación (más nuevo primero)
        containers.sort_by(|a, b| b.1.cmp(&a.1));

        // Eliminar contenedores viejos (mantener solo el más reciente)
        if containers.len() > 1 {
            for (process, _) in containers.iter().skip(1) {
                let timestamp = Local::now().format("%Y-%m-%d %H:%M:%S").to_string();
                
                let log = LogProcess {
                    pid: process.pid,
                    name: process.name.clone(),
                    container_id: process.container_id.clone(),
                    memory_usage: process.memory_usage,
                    cpu_usage: process.cpu_usage,
                    disk_usage_read_mb: process.disk_usage_read_mb,
                    disk_usage_write_mb: process.disk_usage_write_mb,
                    io_read_mb: process.io_read_mb,
                    io_write_mb: process.io_write_mb,
                    action: "killed".to_string(),
                    timestamp,
                };

                log_proc_list.push(log);
                println!("Container killed: ID={}, Name={}", process.container_id, process.name);
                kill_container(process);
            }
        }
    }

    // Enviar logs y datos de RAM
    let timestamp = Local::now().format("%Y-%m-%d %H:%M:%S").to_string();
    log_ram_list.push(LogRam {
        total_ram_mb: system_info.ram_total_mb,
        free_ram_mb: system_info.ram_free_mb,
        usage_ram_mb: system_info.ram_usage_mb,
        timestamp,
    });

    if !log_proc_list.is_empty() {
        let _ = enviar_json_logs(&log_proc_list);
    }
    let _ = enviar_json_ram(&log_ram_list);
}

fn cleanup() {
    // Generar última gráfica
    let _ = get_graph();

    // Eliminar cronjob
    let output = std::process::Command::new("sh")
        .arg("-c")
        .arg("crontab -l | grep -v 'cronJob.sh' | crontab -")
        .output()
        .expect("Failed to remove cron job");
    
    if output.status.success() {
        println!("Cron job removed");
    } else {
        eprintln!("Error removing cron job: {}", String::from_utf8_lossy(&output.stderr));
    }

    // Detener servicios Docker
    let ruta = "../img_docker_logs";
    let output = std::process::Command::new("sh")
        .arg("-c")
        .arg(format!("cd {} && sudo docker-compose down", &ruta))
        .output()
        .expect("Fallo al ejecutar docker compose down");

    if output.status.success() {
        println!("Compose down");
    } else {
        eprintln!(
            "Error al ejecutar Docker Compose: {}",
            String::from_utf8_lossy(&output.stderr)
        );
    }

    println!("------------------------------");
}

fn main() {
    let running = Arc::new(AtomicBool::new(true));
    let r = running.clone();

    ctrlc::set_handler(move || {
        println!("Ctrl+C presionado!");
        cleanup();
        r.store(false, Ordering::SeqCst);
    }).expect("Error al configurar el manejador de Ctrl+C");

    println!("------------------------------");

    let ruta = "../img_docker_logs";
    let mut container_id = String::new();

    let output = std::process::Command::new("sh")
        .arg("-c")
        .arg(format!("cd {} && sudo docker-compose up -d", &ruta))
        .output()
        .expect("Fallo al ejecutar docker compose up");

    if output.status.success() {
        println!("Docker Compose ejecutado correctamente.");

        let id_output = std::process::Command::new("sh")
            .arg("-c")
            .arg("sudo docker ps -q --filter  \"name=logs_container\"")
            .output()
            .expect("Fallo al obtener la ID del contenedor");

        if id_output.status.success() {
            container_id = String::from_utf8_lossy(&id_output.stdout).trim().to_string();
        } else {
            eprintln!(
                "Error al obtener la ID del contenedor: {}",
                String::from_utf8_lossy(&id_output.stderr)
            );
        }
    } else {
        eprintln!(
            "Error al ejecutar Docker Compose: {}",
            String::from_utf8_lossy(&output.stderr)
        );
    }

    println!("------------------------------");

    while running.load(Ordering::SeqCst) {
        let system_info: Result<SystemInfo, _>;
        let json_str = read_proc_file("sysinfo_202044192").unwrap();
        system_info = parse_proc_to_struct(&json_str);

        match system_info {
            Ok(info) => {
                analyzer(info, &container_id);
            }
            Err(e) => println!("Failed to parse JSON: {}", e),
        }

        std::thread::sleep(std::time::Duration::from_secs(10));
    }
}