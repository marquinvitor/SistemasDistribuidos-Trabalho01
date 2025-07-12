
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
    """Mostra o menu de opções para o usuário."""
    """Departamentos Existentes: TI | RH"""
    print("\n--- Gerenciador de Colaboradores ---")
    print("--- Departamentos Existentes: TI | RH ---\n")
    print("1. Adicionar Colaborador")
    print("2. Listar Colaboradores de um Departamento")
    print("3. Calcular Folha Salarial de um Departamento")
    print("4. Demitir Colaborador")
    print("0. Sair")
    return input("Escolha uma opção: ")

def ui_adicionar_colaborador(client: APIClient):
    """Lida com a lógica de pedir os dados de um novo colaborador."""
    nome_depto = input("Digite o nome do departamento: ")
    print("Qual o tipo do colaborador?")
    print("  1. Efetivo")
    print("  2. Autônomo")
    print("  3. Estagiário")
    tipo_escolha = input("Escolha o tipo: ")

    dados = {}
    try:
        dados['id'] = int(input("Digite o ID do colaborador: "))
        dados['nome'] = input("Digite o nome do colaborador: ")
        
        if tipo_escolha == '1':
            dados['tipo'] = 'efetivo'
            dados['salario_mensal'] = float(input("Digite o salário mensal: "))
        elif tipo_escolha == '2':
            dados['tipo'] = 'autonomo'
            dados['horas_trabalhadas'] = int(input("Digite as horas trabalhadas: "))
            dados['valor_hora'] = float(input("Digite o valor por hora: "))
        elif tipo_escolha == '3':
            dados['tipo'] = 'estagiario'
            dados['auxilio_estagio'] = float(input("Digite o valor do auxílio estágio: "))
        else:
            print("Tipo inválido!")
            return
    except ValueError:
        print("Erro: Valor numérico inválido inserido.")
        return

    resultado = client.adicionar_colaborador(nome_depto, dados)
    print("\n>>> Resultado: Colaborador adicionado com sucesso!")

def ui_listar_colaboradores(client: APIClient):
    nome_depto = input("Digite o nome do departamento para listar: ")
    resultado = client.listar_colaboradores(nome_depto)
    if resultado:
        print("\n--- Colaboradores ---")
        print(json.dumps(resultado, indent=2, ensure_ascii=False))
    else:
        print("Nenhum colaborador encontrado ou departamento não existe.")

def ui_calcular_folha(client: APIClient):
    nome_depto = input("Digite o nome do departamento para calcular a folha: ")
    resultado = client.calcular_folha_salarial(nome_depto)
    if resultado:
        print("\n--- Folha Salarial ---")
        print(json.dumps(resultado, indent=2, ensure_ascii=False))
    else:
        print("Não foi possível calcular a folha.")

def ui_demitir_colaborador(client: APIClient):
    nome_depto = input("Digite o nome do departamento: ")
    try:
        colab_id = int(input("Digite o ID do colaborador a ser demitido: "))
    except ValueError:
        print("Erro: ID deve ser um número.")
        return
        
    client.demitir_colaborador(nome_depto, colab_id)
    print(f"\n>>> Tentativa de demissão do colaborador ID {colab_id} enviada.")


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
            print("Saindo do programa. Até mais!")
            break
        else:
            print("Opção inválida. Tente novamente.")
        
        input("\nPressione Enter para continuar...")