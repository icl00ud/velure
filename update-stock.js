#!/usr/bin/env node

const { MongoClient } = require('mongodb');

const uri = "mongodb+srv://claude_user:claude123@velure-cluster-1.6jnzujy.mongodb.net/?appName=Velure-Cluster-1";
const client = new MongoClient(uri);

async function updateProductStock() {
  try {
    await client.connect();
    console.log("✅ Conectado ao MongoDB Atlas");

    const database = client.db('product_service');
    const products = database.collection('products');

    // Primeiro, vamos ver os produtos atuais
    console.log("\n📊 Verificando produtos atuais...");
    const currentProducts = await products.find({}).limit(10).toArray();

    console.log(`\n🔍 Encontrados ${currentProducts.length} produtos (mostrando primeiros 10):`);
    currentProducts.forEach((p, idx) => {
      console.log(`   ${idx + 1}. ${p.name}: quantity = ${p.quantity || 0}`);
    });

    // Contar quantos produtos têm estoque zero
    const zeroStockCount = await products.countDocuments({ $or: [{ quantity: 0 }, { quantity: { $exists: false } }] });
    console.log(`\n⚠️  Produtos com estoque zero ou sem campo quantity: ${zeroStockCount}`);

    // Atualizar todos os produtos com quantidade aleatória entre 15 e 100
    console.log("\n🔄 Atualizando estoque dos produtos...");

    const updateResult = await products.updateMany(
      {},
      [{
        $set: {
          quantity: {
            $add: [
              15,
              { $floor: { $multiply: [{ $rand: {} }, 85] } }
            ]
          }
        }
      }]
    );

    console.log(`✅ ${updateResult.modifiedCount} produtos atualizados!`);

    // Verificar os produtos após atualização
    console.log("\n📊 Verificando produtos após atualização...");
    const updatedProducts = await products.find({}).limit(10).toArray();

    console.log(`\n✨ Produtos atualizados (mostrando primeiros 10):`);
    updatedProducts.forEach((p, idx) => {
      console.log(`   ${idx + 1}. ${p.name}: quantity = ${p.quantity}`);
    });

    // Estatísticas finais
    const totalProducts = await products.countDocuments({});
    const inStockCount = await products.countDocuments({ quantity: { $gt: 0 } });

    console.log(`\n📈 Estatísticas finais:`);
    console.log(`   Total de produtos: ${totalProducts}`);
    console.log(`   Produtos em estoque: ${inStockCount}`);
    console.log(`   Produtos sem estoque: ${totalProducts - inStockCount}`);

  } catch (error) {
    console.error("❌ Erro:", error.message);
    process.exit(1);
  } finally {
    await client.close();
    console.log("\n🔌 Conexão fechada");
  }
}

updateProductStock();
