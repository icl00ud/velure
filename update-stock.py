#!/usr/bin/env python3

from pymongo import MongoClient
from pymongo.server_api import ServerApi
import random

uri = "mongodb+srv://claude_user:claude123@velure-cluster-1.6jnzujy.mongodb.net/?appName=Velure-Cluster-1"

# Create a new client and connect to the server
client = MongoClient(uri, server_api=ServerApi('1'))

try:
    # Send a ping to confirm a successful connection
    client.admin.command('ping')
    print("✅ Conectado ao MongoDB Atlas\n")

    db = client['product_service']
    products_collection = db['products']

    # Verificar produtos atuais
    print("📊 Verificando produtos atuais...")
    current_products = list(products_collection.find({}, {"name": 1, "quantity": 1}).limit(10))

    print(f"\n🔍 Encontrados {products_collection.count_documents({})} produtos (mostrando primeiros 10):")
    for idx, p in enumerate(current_products, 1):
        quantity = p.get('quantity', 0)
        print(f"   {idx}. {p.get('name', 'N/A')}: quantity = {quantity}")

    # Contar quantos produtos têm estoque zero
    zero_stock_count = products_collection.count_documents({
        "$or": [
            {"quantity": 0},
            {"quantity": {"$exists": False}}
        ]
    })
    print(f"\n⚠️  Produtos com estoque zero ou sem campo quantity: {zero_stock_count}")

    # Atualizar todos os produtos com quantidade aleatória entre 15 e 100
    print("\n🔄 Atualizando estoque dos produtos...")

    # Atualizar cada produto com um valor aleatório
    all_products = products_collection.find({})
    updated_count = 0

    for product in all_products:
        new_quantity = random.randint(15, 100)
        products_collection.update_one(
            {"_id": product["_id"]},
            {"$set": {"quantity": new_quantity}}
        )
        updated_count += 1

    print(f"✅ {updated_count} produtos atualizados!")

    # Verificar os produtos após atualização
    print("\n📊 Verificando produtos após atualização...")
    updated_products = list(products_collection.find({}, {"name": 1, "quantity": 1}).limit(10))

    print(f"\n✨ Produtos atualizados (mostrando primeiros 10):")
    for idx, p in enumerate(updated_products, 1):
        print(f"   {idx}. {p.get('name', 'N/A')}: quantity = {p.get('quantity', 0)}")

    # Estatísticas finais
    total_products = products_collection.count_documents({})
    in_stock_count = products_collection.count_documents({"quantity": {"$gt": 0}})

    print(f"\n📈 Estatísticas finais:")
    print(f"   Total de produtos: {total_products}")
    print(f"   Produtos em estoque: {in_stock_count}")
    print(f"   Produtos sem estoque: {total_products - in_stock_count}")

except Exception as e:
    print(f"❌ Erro: {e}")
    exit(1)

finally:
    client.close()
    print("\n🔌 Conexão fechada")
