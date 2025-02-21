#!/bin/sh

# Aguardar o Kafka iniciar
echo "Aguardando o Kafka iniciar..."
sleep 30

# Criar tópicos
echo "Criando tópicos..."
kafka-topics --create --if-not-exists --bootstrap-server kafka:9092 --replication-factor 1 --partitions 1 --topic transacoes
kafka-topics --create --if-not-exists --bootstrap-server kafka:9092 --replication-factor 1 --partitions 1 --topic transacoes_pix
kafka-topics --create --if-not-exists --bootstrap-server kafka:9092 --replication-factor 1 --partitions 1 --topic transacoes_cartao

# Listar tópicos para confirmar
echo "Listando tópicos criados:"
kafka-topics --list --bootstrap-server kafka:9092

echo "Configuração dos tópicos concluída!" 