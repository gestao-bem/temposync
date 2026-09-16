CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    excerpt TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'geral',
    read_min INTEGER NOT NULL DEFAULT 5,
    author TEXT NOT NULL DEFAULT 'Equipe TempoSync',
    published_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS newsletter_subscribers (
    email TEXT PRIMARY KEY NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO posts (slug, title, excerpt, body, category, read_min, author, published_at) VALUES
('tolerancia-5-minutos-clt-art-58',
 'Tolerância de 5 Minutos da CLT (Art. 58): Como Funciona na Prática e Por Que Ela Evita Descontos Injustos',
 'Entenda os detalhes da regra que protege colaborador e empresa em variações acidentais de até 10 minutos diários.',
 'O Artigo 58, parágrafo 1º, da CLT determina que variações de horário no registro de ponto não excedentes a cinco minutos, observado o limite máximo de dez minutos diários, não serão descontadas nem computadas como jornada extraordinária. Na prática, isso significa que uma batida às 8h04 em uma jornada que começa às 8h não é atraso — e uma saída às 17h06 não é hora extra. São minutos de natureza fisiológica e operacional: fila do elevador, trânsito na portaria, conclusão de uma frase antes de encerrar.

A regra funciona como uma faixa de respiro diária, não por batida isolada. Se você acumular 4 minutos pela manhã e 4 minutos à tarde, está dentro do limite de 10 minutos diários. Mas se passar de 10 minutos somados no dia, o excedente inteiro pode ser descontado — é a chamada compensação integral. Muitas empresas aplicam a regra de forma errada, descontando cada minuto além da faixa em uma única batida sem considerar o total do dia.

O TempoSync calcula essa soma automaticamente e avisa antes que você ultrapasse involuntariamente a tolerância. Com o alerta ativo, você recebe uma notificação dez minutos após o horário previsto de saída, evitando tanto o desconto quanto o acúmulo involuntário de horas extras. Para gestores, o cálculo correto da tolerância reduz passivos trabalhistas e elimina disputas no fechamento da folha.',
 'legislacao', 8, 'Lucas Silva', '2026-09-01 09:00:00'),

('guia-portaria-671-rep-p',
 'Guia Definitivo da Portaria MTE nº 671/2021: O que é REP-P e o que mudou no controle eletrônico',
 'Tudo o que colaboradores e líderes precisam saber sobre a validade do espelho de ponto digital com assinatura ICP-Brasil.',
 'A Portaria nº 671/2021 do Ministério do Trabalho e Emprego reorganizou as regras do registro eletrônico de ponto no Brasil. Ela extinguiu o antigo sistema de registro por anotação manual e consolidou duas modalidades: o REP-C, relógio de ponto convencional, e o REP-P, programa de computador ou aplicativo que faz o registro por software. O TempoSync opera como REP-P: registra as batidas com identificação do trabalhador, carimbo de tempo e integridade criptográfica.

O que mudou para quem bate ponto: o comprovante de cada marcação passou a ser obrigatório, o espelho de ponto ganhou validade jurídica com assinatura eletrônica, e as marcações não podem ser alteradas sem rastro de auditoria. Arquivos AFD (registro bruto) e AEJ (espelho consolidado) tornaram-se os formatos oficiais de exportação para fiscalização e para apresentação em processos trabalhistas.

Na prática, isso significa que o seu registro pessoal deixou de depender do RH para ter valor probatório. Com um espelho assinado e hash verificável, você tem em mãos o mesmo documento que a empresa apresentaria ao sindicato ou à fiscalização. O TempoSync gera esse espelho em PDF com hash SHA-256 do conteúdo — qualquer alteração posterior invalida a verificação.',
 'legislacao', 10, 'Dr. Roberto Lima', '2026-08-28 09:00:00'),

('banco-de-horas-vs-hora-extra',
 'Banco de Horas vs. Hora Extra: Como o acúmulo excessivo afeta a recuperação mental no trabalho',
 'Sinais de alerta para identificar sobrecarga antes do esgotamento e boas práticas de compensação semestral.',
 'Banco de horas e hora extra parecem sinônimos, mas produzem efeitos diferentes na sua rotina. A hora extra é paga com adicional de 50% (ou 100% em domingos e feriados); o banco de horas troca o pagamento por tempo de folga futura, desde que exista acordo individual, coletivo ou convenção. O banco dá flexibilidade, mas cobra organização: saldos acumulados vencem — geralmente em seis meses — e, se não compensados, viram pagamento com adicional.

O problema aparece quando o crédito cresce sem controle. Um saldo de 20 ou 30 horas não é conquista: é adiamento de descanso. Estudos de saúde ocupacional mostram que jornadas cronicamente estendidas reduzem a capacidade de recuperação entre dias de trabalho, aumentam o risco cardiovascular e impactam sono e atenção. O sinal clássico é o colaborador que acumula horas durante meses e nunca consegue usá-las porque o time está sempre em pico.

A boa prática é tratar o banco como conta a compensar em ciclos curtos: folgas parciais todo mês, planilhamento das sobras próximas ao vencimento e alerta antes do limite. O TempoSync mostra o saldo por ciclo, a projeção de fechamento e avisa quantos dias restam até o vencimento da CCT — para que o crédito vire descanso, não passivo.',
 'saude', 6, 'Camila Amaral', '2026-08-25 09:00:00'),

('esqueceu-de-bater-o-ponto',
 'Esqueceu de Bater o Ponto? Passo a passo para solicitar ajustes e justificar sem ruídos com o RH',
 'Dicas práticas para manter o espelho de ponto transparente e em dia utilizando comprovantes e registros do dia.',
 'Esquecer uma batida é humano e previsível — o problema é deixar a correção para o fechamento do mês, quando as evidências já esfriaram. A regra de ouro é registrar a solicitação de ajuste no mesmo dia, com o maior número possível de referências: horário aproximado, atividade executada, mensagens trocadas no chat, tempo de reunião ou acesso a sistemas internos.

Quando o ajuste é formalizado no próprio dia, o RH consegue validar com a chefia imediata e o registro entra no espelho sem quebrar a regra de integridade da Portaria 671. Quando fica para depois, a correção exige justificativa escrita mais robusta e pode ser tratada como inconsistência — mesmo sendo um erro honesto.

Um fluxo simples resolve: ao perceber a falha, abra uma solicitação de ajuste com data, horário e motivo; anexe ou cite uma prova (reunião na agenda, pull request, mensagem); e acompanhe o status até a aprovação. No TempoSync, o botão "Solicitar Ajuste/Inclusão" no espelho registra data, horário e motivo, e a solicitação fica rastreável até ser decidida. O espelho nunca é editado diretamente: a correção entra como ajuste auditável, exatamente como exige a norma.',
 'gestao', 4, 'Lucas Silva', '2026-08-21 09:00:00'),

('controle-de-ponto-regime-hibrido',
 'Controle de Ponto no Regime Híbrido: Direitos, deveres e como estabelecer limites saudáveis em casa',
 'A fronteira entre o expediente e a vida pessoal quando o escritório fica na sala de estar e o direito à desconexão.',
 'No regime híbrido, o controle de ponto continua obrigatório — o que muda é a forma. A CLT não exige registro específico para o teletrabalho por produção ou tarefa, mas quando há jornada definida com horário de entrada e saída, o registro é devido. Para o colaborador, isso é proteção: sem registro, não há prova de horas extras, de intervalos suprimidos ou de sobreaviso.

O desafio prático do home office é a fronteira entre trabalho e vida pessoal. O direito à desconexão protege o período fora da jornada: mensagens e chamadas de trabalho fora do horário não podem ser cobradas como disponibilidade permanente. Criar rituais ajuda — horário fixo de encerramento, desligar notificações de trabalho após a saída e manter um espaço físico dedicado ao expediente.

Do lado da gestão, o ponto no híbrido funciona melhor quando é transparente para todos: o time vê a jornada que foi registrada, os intervalos são respeitados e o banco de horas é compensado em prazo. O TempoSync funciona no navegador e no celular com registro por modo de trabalho (home office ou escritório), mantendo a mesma trilha auditável em qualquer lugar.',
 'produtividade', 7, 'Fernanda Faria', '2026-08-18 09:00:00'),

('nr-17-pausas-ergonomicas',
 'NR-17 e as Pausas Ergonômicas: Como pequenas interrupções aumentam a clareza e a precisão',
 'Descubra por que pausas programadas de 10 a 15 minutos não reduzem sua entrega diária e protegem sua saúde osteomuscular.',
 'A Norma Regulamentadora 17 trata da ergonomia no trabalho e estabelece que as pausas devem ser previstas sempre que a atividade exigir. Para trabalho com computador, a recomendação clássica é de pausas curtas e frequentes: a cada 50 minutos de atividade contínua, 10 minutos de descanso, além de pausas de 15 a 20 minutos a cada duas horas. Não são regalias trabalhistas — são prevenção de LER/DORT e fadiga mental.

Pausas bem feitas melhoram a entrega. Estudos de produtividade mostram que interrupções planejadas reduzem erros de atenção sustentada: quem faz pausas curtas mantém a precisão de duas a três horas mais do que quem trabalha em blocos ininterruptos. O intervalo também é o momento de levantar, alongar, beber água e olhar para longe — três gestos simples que reduzem dor cervical e fadiga visual.

O ponto de atenção está na fronteira da CLT: pausas curtas para descanso, como café, não integram a jornada quando previstas em acordo ou convenção (art. 71, §3º). O TempoSync registra pausas rápidas de 15 minutos sem deduzi-las do seu cômputo diário, mantendo o histórico visível no espelho e respeitando o limite legal.',
 'saude', 5, 'Gustavo Leão', '2026-08-14 09:00:00');
