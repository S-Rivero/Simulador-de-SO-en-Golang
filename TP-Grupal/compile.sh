#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

function show_main_menu() {
    echo "==================================="
    echo "       Menu Principal"
    echo "==================================="
    echo "1. Compilar módulo"
    echo "2. Ejecutar pruebas (LOCAL)"
    echo "3. Ejecutar prueba"
    echo "0. Salir"
    echo "==================================="
}

function show_compile_menu() {
    echo "==================================="
    echo "       Compilar modulo"
    echo "==================================="
    echo "1. Compilar CPU"
    echo "2. Compilar MEMORIA"
    echo "3. Compilar KERNEL"
    echo "4. Compilar FILESYSTEM"
    echo "0. Volver"
    echo "==================================="
}

function show_test_menu() {
    echo "==================================="
    echo "       Menu de Pruebas"
    echo "==================================="
    echo "1. Planificación"
    echo "2. Race condition"
    echo "3. Particiones fijas"
    echo "4. Particiones dinámicas"
    echo "5. Fibonacci sequence"
    echo "6. Stress"
    echo "0. Volver"
    echo "==================================="
}

function show_algorithm_menu() {
    echo "==================================="
    echo "    Algoritmos de Planificación"
    echo "==================================="
    echo "1. FIFO"
    echo "2. PRIORIDADES"
    echo "3. CMN"
    echo "0. Volver"
    echo "==================================="
}

function show_module_menu() {
    echo "==================================="
    echo "       Modulo"
    echo "==================================="
    echo "1. CPU"
    echo "2. MEMORIA"
    echo "3. KERNEL"
    echo "4. FILESYSTEM"
    echo "0. Volver"
    echo "==================================="
}

# Función para solicitar una IP con valor por defecto
function ask_ip() {
    local module=$1
    local default_ip="127.0.0.1"
    read -p "Ingrese IP para $module [$default_ip]: " ip
    if [ -z "$ip" ]; then
        echo $default_ip
    else
        echo $ip
    fi
}

# Función para actualizar IPs en un archivo de configuración
function update_config_file() {
    local file=$1
    shift
    local updates=("$@")

    for update in "${updates[@]}"; do
        key=$(echo $update | cut -d= -f1)
        value=$(echo $update | cut -d= -f2)
        # Usar sed para actualizar el archivo de configuración
        sed -i "s/\"$key\": *\"[^\"]*\"/\"$key\": \"$value\"/" "$file"
    done
}

# Función para actualizar todos los archivos de configuración de un módulo
function update_all_configs() {
    local module=$1
    shift
    local updates=("$@")

    # Actualizar config.json principal
    if [ -f "$module/config.json" ]; then
        update_config_file "$module/config.json" "${updates[@]}"
    fi

    # Actualizar todos los archivos en la carpeta configs/
    if [ -d "$module/configs" ]; then
        for config_file in "$module/configs"/*.json; do
            if [ -f "$config_file" ]; then
                update_config_file "$config_file" "${updates[@]}"
            fi
        done
    fi
}

function compile_cpu() {
    echo -e "${GREEN}Configurando CPU...${NC}"
    local memory_ip=$(ask_ip "memoria")
    local kernel_ip=$(ask_ip "kernel")

    update_all_configs "cpu" \
        "ip_memory=$memory_ip" \
        "ip_kernel=$kernel_ip"

    cd cpu && go build && cd ..
    echo -e "${GREEN}CPU compilado exitosamente${NC}"
}

function compile_memory() {
    echo -e "${GREEN}Configurando Memoria...${NC}"
    local kernel_ip=$(ask_ip "kernel")
    local cpu_ip=$(ask_ip "CPU")
    local fs_ip=$(ask_ip "filesystem")

    update_all_configs "memoria" \
        "ip_kernel=$kernel_ip" \
        "ip_cpu=$cpu_ip" \
        "ip_filesystem=$fs_ip"

    cd memoria && go build && cd ..
    echo -e "${GREEN}Memoria compilada exitosamente${NC}"
}

function compile_kernel() {
    echo -e "${GREEN}Configurando Kernel...${NC}"
    local memory_ip=$(ask_ip "memoria")
    local cpu_ip=$(ask_ip "CPU")

    update_all_configs "kernel" \
        "ip_memory=$memory_ip" \
        "ip_cpu=$cpu_ip"

    cd kernel && go build && cd ..
    echo -e "${GREEN}Kernel compilado exitosamente${NC}"
}

function compile_filesystem() {
    echo -e "${GREEN}Configurando Filesystem...${NC}"
    local memory_ip=$(ask_ip "memoria")

    update_all_configs "filesystem" \
        "ip_memory=$memory_ip"

    cd filesystem && go build && cd ..
    echo -e "${GREEN}Filesystem compilado exitosamente${NC}"
}

function compile_module() {
    while true; do
        show_compile_menu
        read -p "Seleccione una opción: " option

        case $option in
            0)
                return
                ;;
            1)
                compile_cpu
                ;;
            2)
                compile_memory
                ;;
            3)
                compile_kernel
                ;;
            4)
                compile_filesystem
                ;;
            *)
                echo -e "${RED}Opción inválida${NC}"
                ;;
        esac
        echo
    done
}

function run_local_tests() {
    gnome-terminal --tab -- bash -c "cd cpu && ./cpu"
    gnome-terminal --tab -- bash -c "cd filesystem && ./filesystem"
    gnome-terminal --tab -- bash -c "cd memoria && ./memoria"
    gnome-terminal --tab -- bash -c "cd kernel && ./kernel THE_EMPTINESS_MACHINE 32 0"
}

function get_test_command() {
    local test_option=$1
    local algo_option=$2
    local module_option=$3

    local test_name=""
    local base_params=""

    case $test_option in
        1)
            test_name="planificacion"
            base_params="PLANI_PROC 32 0"
            ;;
        2)
            test_name="race_condition"
            base_params="RECURSOS_MUTEX_PROC 32 0"
            ;;
        3)
            test_name="particiones_fijas"
            base_params="MEM_FIJA_BASE 12 0"
            ;;
        4)
            test_name="particiones_dinamicas"
            base_params="MEM_DINAMICA_BASE 128 0"
            ;;
        5)
            test_name="fs_fibonacci"
            base_params="PRUEBA_FS 0 0"
            ;;
        6)
            test_name="stress"
            base_params="THE_EMPTINESS_MACHINE 16 0"
            ;;
    esac

    local algo_variant=""
    case $algo_option in
        1) algo_variant="FIFO";;
        2) algo_variant="PRIORIDADES";;
        3) algo_variant="CMN";;
    esac

    local module=""
    case $module_option in
        1) module="cpu";;
        2) module="memoria";;
        3) module="kernel";;
        4) module="filesystem";;
    esac

    if [ -n "$algo_variant" ]; then
        echo -e "\n${GREEN}▶ COMANDO A EJECUTAR:${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}★${NC} ./$module $base_params $test_name $algo_variant ${GREEN}★${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
    else
        echo -e "\n${GREEN}▶ COMANDO A EJECUTAR:${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}★${NC} ./$module $base_params $test_name ${GREEN}★${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
    fi
}

function execute_test() {
    show_test_menu
    read -p "Seleccione una opción: " test_option

    if [ "$test_option" == "0" ]; then
        return
    fi

    local algo_option=0
    if [ "$test_option" == "1" ]; then
        show_algorithm_menu
        read -p "Seleccione una opción: " algo_option
        if [ "$algo_option" == "0" ]; then
            return
        fi
    fi

    show_module_menu
    read -p "Seleccione una opción: " module_option
    if [ "$module_option" == "0" ]; then
        return
    fi

    get_test_command $test_option $algo_option $module_option
}

# Main loop
while true; do
    show_main_menu
    read -p "Seleccione una opción: " option

    case $option in
        0)
            echo "Saliendo..."
            break
            ;;
        1)
            compile_module
            ;;
        2)
            run_local_tests
            ;;
        3)
            execute_test
            ;;
        *)
            echo -e "${RED}Opción inválida${NC}"
            ;;
    esac

    echo
done