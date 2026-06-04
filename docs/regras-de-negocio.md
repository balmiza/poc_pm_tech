# Regras de Negócio — Cadastro de Dispositivos

> Este documento descreve as regras de negócio do serviço de cadastro e
> gerenciamento de dispositivos (celulares) de clientes. Serve como fonte de
> verdade do domínio: o código deve refletir o que está descrito aqui.

## Visão geral do domínio

O serviço gerencia o ciclo de vida de **dispositivos** (celulares) associados a
**clientes**. O objetivo é garantir que apenas dispositivos confirmados por
biometria estejam autorizados a operar em nome de um cliente.

O fluxo principal tem quatro operações: **consultar**, **associar**,
**habilitar** e **cancelar** um dispositivo.

## Entidades

### Cliente

Representa o titular dono do dispositivo. É identificado por um identificador
único (`cliente_id`). O serviço não é dono do cadastro de clientes — assume que
o cliente já existe em uma base externa e apenas referencia seu identificador.

### Dispositivo

Representa o celular de um cliente. Possui:

- Identificador único do dispositivo (`dispositivo_id`)
- Referência ao cliente dono (`cliente_id`)
- Estado atual no ciclo de vida (ver seção "Estados do dispositivo")
- Data de associação
- Data de habilitação (quando aplicável)
- Data de cancelamento (quando aplicável)

## Estados do dispositivo

Um dispositivo passa por três estados ao longo do seu ciclo de vida:

| Estado | Significado |
|--------|-------------|
| `pendente_confirmacao` | Dispositivo associado ao cliente, aguardando confirmação por biometria. Ainda **não está autorizado** a operar. |
| `ativo` | Dispositivo confirmado por biometria e habilitado. **Autorizado** a operar. |
| `cancelado` | Dispositivo descadastrado pelo cliente. Não está autorizado e não pode voltar a ficar ativo. |

As transições de estado permitidas são:

- `pendente_confirmacao` → `ativo` (via habilitação, após biometria OK)
- `pendente_confirmacao` → `cancelado` (via cancelamento)
- `ativo` → `cancelado` (via cancelamento)

Não existe transição de volta. Um dispositivo `cancelado` é terminal: para usar
o mesmo aparelho novamente, é preciso fazer uma nova associação (novo registro).

## Regra: um dispositivo ativo por cliente

Um cliente pode ter **no máximo um dispositivo ativo por vez**.

Isso implica que, para associar um novo dispositivo, o cliente não pode ter outro
dispositivo em estado `ativo` ou `pendente_confirmacao`. O dispositivo anterior
precisa estar `cancelado` antes de iniciar uma nova associação.

> Esta é uma restrição inicial da POC. Suporte a múltiplos dispositivos ativos
> está fora do escopo atual.

## Operação: consultar autorização do dispositivo

Permite verificar se um cliente possui um dispositivo autorizado.

- **Entrada:** identificador do cliente.
- **Saída:** indica se o cliente possui um dispositivo no estado `ativo`.
- Um dispositivo só é considerado autorizado quando está no estado `ativo`.
  Dispositivos em `pendente_confirmacao` ou `cancelado` não são autorizados.

Esta é a operação mais consultada do serviço, pois é usada por outros sistemas
para decidir se permitem uma operação sensível em nome do cliente.

## Operação: associar dispositivo

Primeira etapa do cadastro. Vincula um dispositivo a um cliente.

- **Entrada:** identificador do cliente e dados do dispositivo.
- **Resultado:** cria um dispositivo no estado `pendente_confirmacao`.
- **Não exige biometria nesta etapa.** A associação apenas registra a intenção
  de cadastrar o dispositivo.
- **Pré-condição:** o cliente não pode ter outro dispositivo `ativo` ou
  `pendente_confirmacao` (ver "Regra: um dispositivo ativo por cliente").

## Operação: habilitar dispositivo

Segunda etapa do cadastro. Confirma a identidade do cliente e ativa o dispositivo.

- **Pré-condição:** o dispositivo precisa estar no estado `pendente_confirmacao`.
- **Validação de biometria:** antes de habilitar, o serviço consulta uma **API
  externa de biometria facial** (ver "Integração: biometria facial") para
  confirmar a identidade do cliente.
- **Resultado:**
  - Se a biometria for aprovada, o dispositivo passa para `ativo`.
  - Se a biometria for reprovada, o dispositivo permanece em
    `pendente_confirmacao` e a habilitação não acontece.

## Operação: cancelar dispositivo

Permite ao cliente descadastrar seu dispositivo.

- **Tipo:** cancelamento lógico (*soft-delete*). O registro **não é removido** da
  base.
- **Resultado:** o dispositivo passa para o estado `cancelado` e a data de
  cancelamento é registrada.
- **Pré-condição:** o dispositivo precisa estar em `pendente_confirmacao` ou
  `ativo`. Cancelar um dispositivo já `cancelado` não tem efeito.
- Manter o histórico é intencional: precisamos saber quais dispositivos um
  cliente já teve e quando foram cancelados.

## Integração: biometria facial

A validação de biometria facial é feita por uma **API externa, fora do domínio
deste serviço**.

- O serviço **não armazena** nenhum dado biométrico (nem imagem, nem template).
- O serviço **apenas consulta** a API externa, passando o identificador do
  cliente, e usa a resposta (aprovado / reprovado) para decidir se habilita o
  dispositivo.
- A biometria é tratada como uma dependência externa transparente: este serviço
  confia no resultado retornado e não tem responsabilidade sobre como a
  verificação é feita.

## Glossário de termos

- **Dispositivo:** o celular de um cliente, gerenciado por este serviço.
- **Associar:** primeira etapa do cadastro; cria o dispositivo em
  `pendente_confirmacao`, sem biometria.
- **Habilitar:** segunda etapa do cadastro; após biometria aprovada, ativa o
  dispositivo.
- **Autorizado:** dispositivo no estado `ativo`, apto a operar em nome do cliente.
- **Cancelamento lógico (soft-delete):** mudança de estado para `cancelado` sem
  remover o registro da base.
- **Biometria facial:** verificação de identidade feita por API externa; não é
  responsabilidade deste serviço.