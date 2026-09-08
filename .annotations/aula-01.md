`Go` não possui construtores como parte da linguagem. Chamamos de construtor uma função como `New(...)` que cria um valor somente depois de verificar suas regras.


Como name começa com letra minúscula, somente o pacote event pode alterá-lo diretamente.

Sem a tag:

type HealthResponse struct {
    Status string
}

O JSON normalmente seria:

{
"Status": "ok"
}

Com a tag:

type HealthResponse struct {
    Status string `json:"status"`
}

O resultado será:

{
"status": "ok"
}

 Importante: a tag não valida o valor e não altera o nome do campo dentro do Go.
  Ela só controla sua representação no JSON.

---

## Questões respondidas
1. **Por que validar no handler e no construtor?**
O handler valida questões de transporte: JSON malformado, campo ausente, data que não pode ser decodificada.

O domínio valida regras de negócio: capacidade positiva, preço não negativo, vendas antes do evento.

*Se as regras existissem apenas no handler, um worker, CLI ou outro serviço poderia criar um evento inválido diretamente. O construtor garante que qualquer caminho de criação respeite as invariantes.*

2. **Por que não usar float64 para dinheiro?**
Números decimais como 0.1 normalmente não possuem representação binária exata. Erros pequenos podem aparecer depois de somas, multiplicações e arredondamentos. Representar R$ 145,00 como 14500 centavos em int64 mantém cálculos inteiros e previsíveis.

3. **O que campos privados protegem?**
Em Go, não temos classes; a visibilidade é controlada por pacote. 

Qualquer código dentro do pacote `event` consegue alterar esses campos diretamente. Código de outro pacote não consegue. Isso reduz alterações acidentais, mas não impede que um método mal implementado do próprio pacote crie estado inválido.

4. **Por que usar datas fixas nos testes?**
Para tornar os testes determinísticos: as mesmas entradas sempre produzem o mesmo resultado.

`time.Now()` muda a cada execução e pode gerar falhas internitentes perto de segundos, viradas de dia ou mudanças de fuso. Nest domínio só precisamos comparar a ordem entre duas datas, portanto náo existe motivo para consultar o relógio real.