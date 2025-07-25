import requests
import json

class APIClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url

    def _make_request(self, method, endpoint, **kwargs):
        url = f"{self.base_url}/{endpoint}"
        try:
            response = requests.request(method, url, **kwargs)
            response.raise_for_status()
            if response.status_code == 204:
                return {"success": True, "message": "Operação realizada com sucesso."}
            if response.text:
                return response.json()
            return None
        except requests.exceptions.RequestException as e:
            return {"success": False, "error": str(e)}

    def adicionar_colaborador(self, nome_depto: str, dados: dict):
        return self._make_request("POST", f"departamentos/{nome_depto}/colaboradores", json=dados)

    def listar_colaboradores(self, nome_depto: str):
        return self._make_request("GET", f"departamentos/{nome_depto}/colaboradores")

    def calcular_folha_salarial(self, nome_depto: str):
        return self._make_request("GET", f"departamentos/{nome_depto}/folha-salarial")

    def demitir_colaborador(self, nome_depto: str, id: int):
        return self._make_request("DELETE", f"departamentos/{nome_depto}/colaboradores/{id}")


def exibir_menu():
    print("\n--- Gerenciador de Colaboradores ---")
    print("Departamentos disponíveis: TI | RH\n")
    print("1. Adicionar Colaborador")
    print("2. Listar Colaboradores de um Departamento")
    print("3. Calcular Folha Salarial de um Departamento")
    print("4. Demitir Colaborador")
    print("0. Sair\n")
    return input("Escolha uma opção: ")


def ui_adicionar_colaborador(client: APIClient):
    print("\n--- Adicionar Colaborador ---\n")
    nome_depto = input("Departamento: ")

    print("\nTipo do colaborador:")
    print("  1. Efetivo")
    print("  2. Autônomo")
    print("  3. Estagiário")
    tipo_escolha = input("Escolha o tipo (1-3): ")

    dados = {}
    try:
        dados['id'] = int(input("\nID do colaborador: "))
        dados['nome'] = input("Nome do colaborador: ")

        if tipo_escolha == '1':
            dados['tipo'] = 'efetivo'
            dados['salario_mensal'] = float(input("Salário mensal: "))
        elif tipo_escolha == '2':
            dados['tipo'] = 'autonomo'
            dados['horas_trabalhadas'] = int(input("Horas trabalhadas: "))
            dados['valor_hora'] = float(input("Valor por hora: "))
        elif tipo_escolha == '3':
            dados['tipo'] = 'estagiario'
            dados['auxilio_estagio'] = float(input("Valor do auxílio estágio: "))
        else:
            print("\nTipo inválido.")
            return
    except ValueError:
        print("\nErro: valor numérico inválido.")
        return

    resultado = client.adicionar_colaborador(nome_depto, dados)
    print("\nColaborador adicionado com sucesso.\n")


def ui_listar_colaboradores(client: APIClient):
    print("\n--- Listar Colaboradores ---\n")
    nome_depto = input("Departamento: ")
    resultado = client.listar_colaboradores(nome_depto)

    if resultado:
        print("\nColaboradores:\n")
        print(json.dumps(resultado, indent=2, ensure_ascii=False))
    else:
        print("\nNenhum colaborador encontrado ou departamento inexistente.\n")


def ui_calcular_folha(client: APIClient):
    print("\n--- Calcular Folha Salarial ---\n")
    nome_depto = input("Departamento: ")
    resultado = client.calcular_folha_salarial(nome_depto)

    if resultado:
        print("\nFolha Salarial:\n")
        print(json.dumps(resultado, indent=2, ensure_ascii=False))
    else:
        print("\nNão foi possível calcular a folha.\n")


def ui_demitir_colaborador(client: APIClient):
    print("\n--- Demitir Colaborador ---\n")
    nome_depto = input("Departamento: ")
    try:
        colab_id = int(input("ID do colaborador a ser demitido: "))
    except ValueError:
        print("\nErro: ID deve ser um número.")
        return

    resultado = client.demitir_colaborador(nome_depto, colab_id)
    print(f"\nResultado: {resultado.get('message', 'Solicitação enviada.')}\n")


if __name__ == "__main__":
    api_client = APIClient()

    while True:
        escolha = exibir_menu()

        if escolha == '1':
            ui_adicionar_colaborador(api_client)
        elif escolha == '2':
            ui_listar_colaboradores(api_client)
        elif escolha == '3':
            ui_calcular_folha(api_client)
        elif escolha == '4':
            ui_demitir_colaborador(api_client)
        elif escolha == '0':
            print("\nSaindo do programa.\n")
            break
        else:
            print("\nOpção inválida.\n")

        input("Pressione Enter para continuar...")
