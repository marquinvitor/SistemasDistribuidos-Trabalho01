const inquirer = require("inquirer");
const fetch = require("node-fetch");

const BASE_URL = "http://localhost:8080";

async function adicionarColaborador() {
  const { departamento, tipo } = await inquirer.prompt([
    {
      name: "departamento",
      message: "Nome do departamento:",
      type: "input",
    },
    {
      name: "tipo",
      message: "Tipo do colaborador:",
      type: "list",
      choices: ["efetivo", "autonomo", "estagiario"],
    },
  ]);

  const respostasBase = await inquirer.prompt([
    { name: "id", message: "ID do colaborador:", type: "number" },
    { name: "nome", message: "Nome do colaborador:" },
  ]);

  let payload = {
    tipo,
    id: respostasBase.id,
    nome: respostasBase.nome,
  };

  if (tipo === "efetivo") {
    const { salario_mensal } = await inquirer.prompt([
      { name: "salario_mensal", message: "Salário mensal:", type: "number" },
    ]);
    payload.salario_mensal = salario_mensal;
  } else if (tipo === "autonomo") {
    const { horas_trabalhadas, valor_hora } = await inquirer.prompt([
      {
        name: "horas_trabalhadas",
        message: "Horas trabalhadas:",
        type: "number",
      },
      { name: "valor_hora", message: "Valor por hora:", type: "number" },
    ]);
    payload.horas_trabalhadas = horas_trabalhadas;
    payload.valor_hora = valor_hora;
  } else if (tipo === "estagiario") {
    const { auxilio_estagio } = await inquirer.prompt([
      { name: "auxilio_estagio", message: "Auxílio estágio:", type: "number" },
    ]);
    payload.auxilio_estagio = auxilio_estagio;
  }

  const res = await fetch(
    `${BASE_URL}/departamentos/${departamento}/colaboradores`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }
  );

  const text = await res.text();
  console.log(`\nResposta: ${text}\n`);
}

async function listarColaboradores() {
  const { departamento } = await inquirer.prompt([
    { name: "departamento", message: "Nome do departamento:", type: "input" },
  ]);

  const res = await fetch(
    `${BASE_URL}/departamentos/${departamento}/colaboradores`
  );
  const json = await res.json();
  console.log(`\nColaboradores:\n`, json);
}

async function calcularFolha() {
  const { departamento } = await inquirer.prompt([
    { name: "departamento", message: "Nome do departamento:", type: "input" },
  ]);

  const res = await fetch(
    `${BASE_URL}/departamentos/${departamento}/folha-salarial`
  );
  const json = await res.json();
  console.log(`\n Folha salarial:\n`, json);
}

async function demitirColaborador() {
  const { departamento, id } = await inquirer.prompt([
    { name: "departamento", message: "Nome do departamento:", type: "input" },
    { name: "id", message: "ID do colaborador a demitir:", type: "number" },
  ]);

  const res = await fetch(
    `${BASE_URL}/departamentos/${departamento}/colaboradores/${id}`,
    {
      method: "DELETE",
    }
  );

  const text = await res.text();
  console.log(`\nResposta: ${text}`);
}

async function menu() {
  while (true) {
    console.log("Departamentos disponíveis: TI | RH\n");
    const { escolha } = await inquirer.prompt([
      {
        name: "escolha",
        message: "Escolha uma opção:",
        type: "list",
        choices: [
          { name: "1. Adicionar colaborador", value: "add" },
          { name: "2. Listar colaboradores", value: "list" },
          { name: "3. Calcular folha salarial", value: "folha" },
          { name: "4. Demitir colaborador", value: "fire" },
          { name: "0. Sair", value: "exit" },
        ],
      },
    ]);

    switch (escolha) {
      case "add":
        await adicionarColaborador();
        break;
      case "list":
        await listarColaboradores();
        break;
      case "folha":
        await calcularFolha();
        break;
      case "fire":
        await demitirColaborador();
        break;
      case "exit":
        console.log("Saindo...");
        return;
    }

    console.log("──────────────────────────────\n");
  }
}

menu();
