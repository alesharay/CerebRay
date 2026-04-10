-- Cerebray seed data for local development
-- Idempotent: safe to run multiple times (uses ON CONFLICT DO NOTHING)

BEGIN;

-- ---------------------------------------------------------------------------
-- User
-- ---------------------------------------------------------------------------
INSERT INTO users (id, oidc_subject, email, name)
VALUES (1, 'local', 'admin@localhost', 'Local Admin')
ON CONFLICT (oidc_subject) DO NOTHING;

SELECT setval('users_id_seq', GREATEST(1, (SELECT MAX(id) FROM users)));

-- ---------------------------------------------------------------------------
-- Conversation (inserted before notes so source_chat_id FK works)
-- ---------------------------------------------------------------------------
INSERT INTO conversations (id, user_id, title, topic)
VALUES (1, 1, 'Learning Design Patterns', 'software design patterns')
ON CONFLICT DO NOTHING;

SELECT setval('conversations_id_seq', GREATEST(1, (SELECT MAX(id) FROM conversations)));

-- Messages for the conversation
INSERT INTO messages (id, conversation_id, role, content, model) VALUES
(1, 1, 'user',      'Can you explain the Observer pattern to me?', ''),
(2, 1, 'assistant', 'The Observer pattern defines a one-to-many dependency between objects. When one object (the subject) changes state, all its dependents (observers) get notified and updated automatically. Think of it like subscribing to a newsletter - you register your interest, and updates come to you.', 'claude-sonnet-4-20250514'),
(3, 1, 'user',      'How does it compare to the Pub/Sub pattern?', ''),
(4, 1, 'assistant', 'Good question. Observer and Pub/Sub are related but differ in coupling. In Observer, the subject knows about its observers directly. In Pub/Sub, a message broker sits in between - publishers and subscribers never know about each other. Pub/Sub scales better across distributed systems, while Observer works well within a single application.', 'claude-sonnet-4-20250514'),
(5, 1, 'user',      'When would I pick one over the other?', ''),
(6, 1, 'assistant', 'Use Observer when your components live in the same process and you want simple, direct notification - UI frameworks rely on this heavily. Choose Pub/Sub when you need decoupling across service boundaries, want to add subscribers without touching existing code, or need message persistence and replay. A rule of thumb: Observer for in-process, Pub/Sub for cross-process.', 'claude-sonnet-4-20250514')
ON CONFLICT DO NOTHING;

SELECT setval('messages_id_seq', GREATEST(6, (SELECT MAX(id) FROM messages)));

-- ---------------------------------------------------------------------------
-- Notes
-- ---------------------------------------------------------------------------
INSERT INTO notes (id, user_id, title, summary, laymans_terms, analogy, core_idea, body, components, why_it_matters, examples, note_type, status, tlp, source_chat_id) VALUES

-- Fleeting notes (inbox, minimal content)
(1,  1, 'Look into CRDT data structures', '', '', '', '', 'Heard about CRDTs in a distributed systems talk. Might be useful for offline-first apps.', '', '', '', 'concept', 'fleeting', 'clear', NULL),
(2,  1, 'Spaced repetition for code review', '', '', '', '', '', '', '', '', 'insight', 'fleeting', 'clear', NULL),
(3,  1, 'PostgreSQL partial indexes', '', '', '', '', 'Can create indexes that only cover rows matching a WHERE clause. Good for status-based queries.', '', '', '', 'concept', 'fleeting', 'clear', NULL),
(4,  1, 'Category theory in FP', '', '', '', '', 'Functors, monads, applicatives - how do these map to practical programming?', '', '', '', 'question', 'fleeting', 'clear', NULL),

-- Sleeping notes (partially filled out)
(5,  1, 'Observer Pattern',
     'A behavioral design pattern where objects subscribe to state changes in another object.',
     '', '', 'Defines a one-to-many dependency so that when one object changes state, all dependents are notified.',
     '', '', '', '', 'concept', 'sleeping', 'clear', 1),
(6,  1, 'Dependency Injection',
     'A technique where an object receives its dependencies from the outside rather than creating them internally.',
     '', '', '',
     'Instead of a class instantiating its own collaborators, they get passed in through constructors or setters. This makes testing easier and reduces coupling.',
     '', '', '', 'concept', 'sleeping', 'clear', NULL),
(7,  1, 'Bloom Filters',
     'A probabilistic data structure for set membership queries.',
     '', '', '',
     'Bloom filters can tell you "definitely not in set" or "probably in set." They use multiple hash functions mapping to a bit array. False positives are possible, false negatives are not.',
     '', '', '', 'concept', 'sleeping', 'clear', NULL),

-- Active/linked notes (fully fleshed out)
(8,  1, 'Strategy Pattern',
     'Defines a family of algorithms, encapsulates each one, and makes them interchangeable at runtime.',
     'Imagine a GPS app that can switch between "fastest route," "shortest route," and "avoid highways" without changing the navigation screen itself.',
     'A toolbox where you swap out the specific tool (sorting algorithm, payment processor) without redesigning the workbench.',
     'Separate what varies (the algorithm) from what stays the same (the client code) by programming to an interface.',
     'The Strategy pattern pulls an algorithm out of the class that uses it and puts it behind an interface. The client holds a reference to the interface and delegates work to whatever concrete strategy is plugged in.',
     '- **Context**: the class that delegates to a strategy\n- **Strategy interface**: the contract all algorithms share\n- **Concrete strategies**: individual algorithm implementations',
     'Eliminates conditional blocks that pick behavior based on type. Makes it trivial to add new algorithms without touching existing code.',
     '- Sorting: swap between quicksort, mergesort, heapsort\n- Payment: swap between Stripe, PayPal, bank transfer\n- Compression: swap between gzip, brotli, zstd',
     'concept', 'active', 'clear', NULL),

(9,  1, 'Single Responsibility Principle',
     'A class should have only one reason to change.',
     'A cook should focus on cooking, not also managing the restaurant accounts.',
     'Each worker on an assembly line has one specific job. If you need to change how painting works, you only retrain the painter.',
     'Group together the things that change for the same reason. Separate things that change for different reasons.',
     'SRP is the first of the SOLID principles. It does not mean a class should only do one thing - it means a class should only be responsible to one actor or stakeholder. If two different departments would request changes to the same class, that class has too many responsibilities.',
     '- Identify the actors (stakeholders) that might request changes\n- Each class should serve exactly one actor\n- If a method serves a different actor, extract it into its own class',
     'Reduces merge conflicts in teams. Makes classes easier to test because each has a focused purpose. Limits blast radius when requirements change.',
     '- A `Report` class that both formats and calculates data violates SRP\n- Split into `ReportCalculator` and `ReportFormatter`',
     'theory', 'linked', 'clear', NULL),

(10, 1, 'Composition Over Inheritance',
     'Favor assembling behavior from small, focused components rather than building deep inheritance hierarchies.',
     'Instead of buying a Swiss Army knife that does everything (but nothing well), pick individual tools that each do one job perfectly.',
     'Building with LEGO bricks vs. carving from a single block of wood. Bricks snap together in new ways; a carving is fixed.',
     'Achieve code reuse and polymorphism by composing objects with well-defined interfaces rather than inheriting from base classes.',
     'Deep inheritance trees create tight coupling and fragile hierarchies. Changing a base class can break dozens of subclasses. Composition lets you mix and match behaviors through interfaces and delegation, keeping each piece independent and replaceable.',
     '- Define small interfaces for each capability\n- Implement capabilities as standalone types\n- Inject them into the struct/class that needs them',
     'Go enforces this by design - no inheritance at all, only interfaces and embedding. Many experienced OOP developers converge on this pattern after fighting brittle hierarchies.',
     '- Go''s `io.Reader`/`io.Writer` interfaces\n- React''s component composition vs. class inheritance\n- Entity Component System in game engines',
     'theory', 'linked', 'clear', NULL),

(11, 1, 'Pub/Sub vs Observer',
     'Observer couples subject to observers directly. Pub/Sub decouples them through a message broker.',
     'Observer is shouting across the room to specific people. Pub/Sub is posting on a bulletin board that anyone can read.',
     'Observer is a phone call (direct connection). Pub/Sub is a radio broadcast (indirect, one-to-many, no knowledge of listeners).',
     'The key difference is coupling: Observer is synchronous and direct, Pub/Sub is asynchronous and brokered.',
     'Both patterns handle event-driven communication, but they sit at different points on the coupling spectrum. Observer works within a single process where the subject maintains a list of observers. Pub/Sub introduces a broker (message queue, event bus) that fully decouples producers from consumers.',
     '- **Observer**: subject holds observer references, calls them directly\n- **Pub/Sub**: publisher sends messages to a topic, broker delivers to subscribers\n- **Coupling**: Observer = tight, Pub/Sub = loose',
     'Choosing the wrong pattern leads to either unnecessary infrastructure (Pub/Sub for simple UI updates) or spaghetti coupling (Observer across microservices).',
     '- Observer: React useState/useEffect, DOM event listeners\n- Pub/Sub: Kafka, RabbitMQ, Redis Pub/Sub, AWS SNS',
     'insight', 'linked', 'clear', 1),

(12, 1, 'Zettelkasten Method',
     'A note-taking system built on atomic, interconnected notes rather than hierarchical folders.',
     'Instead of filing papers into folders, you write one idea per card and link cards together with references. The links are more valuable than any folder structure.',
     'A wiki for your brain. Each page is one idea, and the hyperlinks between pages form your understanding.',
     'Write atomic notes. Connect them. Let structure emerge from connections rather than imposing categories up front.',
     'Niklas Luhmann developed the Zettelkasten ("slip box") and used it to write over 70 books. The system works because it mirrors how knowledge actually grows: through connections between ideas, not filing into predetermined buckets. Each note should be self-contained and express one idea clearly enough that a future reader (your future self) can understand it without context.',
     '- Fleeting notes: quick captures, inbox items\n- Literature notes: summaries from sources\n- Permanent notes: refined, atomic ideas in your own words\n- Connections: links between notes with context on why they relate',
     'Combats the collector''s fallacy (saving information without processing it). Forces you to think through ideas and relate them to what you already know.',
     '- This app (Cerebray) implements digital Zettelkasten\n- Obsidian, Roam Research, and Logseq are popular Zettelkasten tools',
     'guide', 'active', 'green', NULL),

(13, 1, 'Retrieval Practice',
     'Testing yourself on material is more effective for long-term retention than re-reading or highlighting.',
     'You learn more by trying to recall the answer before looking it up than by reading the answer ten times.',
     'A muscle gets stronger from lifting weights (active use), not from watching someone else lift (passive observation).',
     'Actively pulling information from memory strengthens the neural pathways to that information.',
     'Research consistently shows that the act of retrieving information from memory strengthens that memory far more than passive review. This is sometimes called the "testing effect." It works because retrieval is a form of elaborative encoding - your brain must reconstruct the knowledge, which reinforces connections.',
     '- Free recall: write down everything you remember about a topic\n- Flashcards: classic spaced retrieval tool\n- Practice problems: apply knowledge rather than just recognize it\n- Teaching: explaining to someone else forces retrieval and reveals gaps',
     'Most people default to re-reading, which feels productive but creates an illusion of competence. Retrieval practice is harder and feels less comfortable, which is exactly why it works.',
     '- Anki flashcard system\n- Feynman Technique (explain it simply)\n- Cornell note-taking method (cover the notes, recall from cues)',
     'theory', 'active', 'clear', NULL),

(14, 1, 'Interleaving Practice',
     'Mixing different topics or problem types during study sessions improves learning compared to blocked practice.',
     'Instead of practicing 20 addition problems, then 20 subtraction problems, shuffle them together.',
     'Cross-training in sports. Runners who also swim and cycle perform better overall than those who only run.',
     'Interleaving forces your brain to discriminate between problem types and select the right strategy, which is the real skill.',
     'Blocked practice (doing the same type of problem repeatedly) feels easier and produces faster short-term gains. Interleaving feels harder and slower, but the long-term retention and transfer is significantly better. The difficulty is "desirable" because it forces deeper processing.',
     '- Mix problem types within a single study session\n- Alternate between related but distinct topics\n- Space the interleaving across multiple days',
     'Particularly important for fields where the challenge is not executing a known procedure but recognizing which procedure to apply.',
     '- Study math by mixing algebra, geometry, and statistics problems\n- Practice programming by alternating between different data structures\n- Study music by mixing scales, arpeggios, and pieces',
     'theory', 'linked', 'clear', NULL),

(15, 1, 'CAP Theorem',
     'A distributed system can provide at most two of three guarantees: Consistency, Availability, and Partition tolerance.',
     'You can have a system that is always correct, always responsive, or always handles network splits - but not all three at once.',
     'A restaurant with multiple kitchens. If the phone line between kitchens goes down (partition), you either stop taking orders (sacrifice availability) or risk serving inconsistent dishes (sacrifice consistency).',
     'In the presence of a network partition, you must choose between consistency and availability.',
     'Eric Brewer''s CAP theorem (proven by Gilbert and Lynch in 2002) states that when a network partition occurs, a distributed system must make a tradeoff. CP systems (like ZooKeeper) refuse to respond if they cannot guarantee the latest data. AP systems (like Cassandra) always respond but might return stale data. In practice, partitions are rare but inevitable, so the real question is: what does your system do when one happens?',
     '- **Consistency**: every read returns the most recent write\n- **Availability**: every request gets a response (not an error)\n- **Partition tolerance**: the system works despite network splits between nodes',
     'Drives architectural decisions in distributed systems. Choosing the wrong tradeoff leads to data loss (AP when you needed CP) or downtime (CP when you needed AP).',
     '- CP: ZooKeeper, HBase, MongoDB (default config)\n- AP: Cassandra, DynamoDB, CouchDB\n- CA: only possible on a single node (no distribution)',
     'concept', 'linked', 'clear', NULL),

-- Archived notes
(16, 1, 'MVC Pattern Basics',
     'Model-View-Controller separates data, presentation, and input handling.',
     '', '', 'Divide an application into three roles: the data and logic (Model), the display (View), and the input handler (Controller).',
     'Classic architectural pattern from the 1970s. Still used but largely superseded by variations like MVVM, MVP, and component-based architectures in modern frameworks.',
     '', '', '',
     'structure', 'archived', 'clear', NULL),

(17, 1, 'Waterfall vs Agile',
     'Waterfall is sequential; Agile is iterative. Most teams now use Agile or hybrid approaches.',
     '', '', '',
     'Captured from a project management overview. Not deeply relevant to my current technical focus.',
     '', '', '',
     'reference', 'archived', 'clear', NULL)

ON CONFLICT DO NOTHING;

SELECT setval('notes_id_seq', GREATEST(17, (SELECT MAX(id) FROM notes)));

-- ---------------------------------------------------------------------------
-- Tags
-- ---------------------------------------------------------------------------
INSERT INTO tags (id, user_id, name) VALUES
(1,  1, 'programming'),
(2,  1, 'learning'),
(3,  1, 'design-patterns'),
(4,  1, 'algorithms'),
(5,  1, 'architecture'),
(6,  1, 'testing'),
(7,  1, 'databases'),
(8,  1, 'networking'),
(9,  1, 'distributed-systems'),
(10, 1, 'study-techniques')
ON CONFLICT DO NOTHING;

SELECT setval('tags_id_seq', GREATEST(10, (SELECT MAX(id) FROM tags)));

-- ---------------------------------------------------------------------------
-- Note-tag associations
-- ---------------------------------------------------------------------------
INSERT INTO note_tags (note_id, tag_id) VALUES
-- Fleeting notes
(1,  9),   -- CRDTs > distributed-systems
(1,  7),   -- CRDTs > databases
(2,  2),   -- Spaced repetition > learning
(2,  6),   -- Spaced repetition > testing
(3,  7),   -- Partial indexes > databases
(4,  1),   -- Category theory > programming

-- Sleeping notes
(5,  3),   -- Observer > design-patterns
(5,  1),   -- Observer > programming
(6,  3),   -- DI > design-patterns
(6,  6),   -- DI > testing
(7,  4),   -- Bloom filters > algorithms
(7,  7),   -- Bloom filters > databases

-- Active/linked notes
(8,  3),   -- Strategy > design-patterns
(8,  1),   -- Strategy > programming
(9,  3),   -- SRP > design-patterns
(9,  5),   -- SRP > architecture
(10, 3),   -- Composition > design-patterns
(10, 5),   -- Composition > architecture
(11, 3),   -- Pub/Sub vs Observer > design-patterns
(11, 5),   -- Pub/Sub vs Observer > architecture
(12, 2),   -- Zettelkasten > learning
(13, 2),   -- Retrieval practice > learning
(13, 10),  -- Retrieval practice > study-techniques
(14, 2),   -- Interleaving > learning
(14, 10),  -- Interleaving > study-techniques
(15, 9),   -- CAP > distributed-systems
(15, 5),   -- CAP > architecture

-- Archived notes
(16, 3),   -- MVC > design-patterns
(16, 5),   -- MVC > architecture
(17, 5)    -- Waterfall vs Agile > architecture
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------------
-- Connections between notes
-- ---------------------------------------------------------------------------
INSERT INTO connections (id, source_id, target_id, label) VALUES
(1,  5,  11, 'compared in'),           -- Observer > Pub/Sub vs Observer
(2,  11, 5,  'analyzes'),              -- Pub/Sub vs Observer > Observer
(3,  8,  9,  'relates to'),            -- Strategy > SRP
(4,  9,  10, 'builds on'),             -- SRP > Composition
(5,  10, 9,  'complements'),           -- Composition > SRP
(6,  8,  6,  'often used with'),       -- Strategy > DI
(7,  13, 14, 'complements'),           -- Retrieval practice > Interleaving
(8,  14, 13, 'complements'),           -- Interleaving > Retrieval practice
(9,  12, 13, 'technique for'),         -- Zettelkasten > Retrieval practice
(10, 15, 1,  'example of'),            -- CAP > CRDTs (fleeting)
(11, 16, 8,  'superseded by'),         -- MVC > Strategy (different approach)
(12, 5,  8,  'contrasts with'),        -- Observer > Strategy (behavioral patterns)
(13, 11, 15, 'relates to')             -- Pub/Sub vs Observer > CAP
ON CONFLICT DO NOTHING;

SELECT setval('connections_id_seq', GREATEST(13, (SELECT MAX(id) FROM connections)));

-- ---------------------------------------------------------------------------
-- Glossary terms
-- ---------------------------------------------------------------------------
INSERT INTO glossary_terms (id, user_id, term, definition, source_note_id) VALUES
(1, 1, 'Design Pattern',     'A reusable solution to a commonly occurring problem in software design. Not a finished design, but a template for how to solve a problem in many different situations.', 8),
(2, 1, 'SOLID',              'Five design principles (Single Responsibility, Open-Closed, Liskov Substitution, Interface Segregation, Dependency Inversion) that make object-oriented designs more maintainable.', 9),
(3, 1, 'Zettelkasten',       'German for "slip box." A knowledge management method based on atomic, interconnected notes rather than hierarchical filing.', 12),
(4, 1, 'CAP Theorem',        'Brewer''s theorem stating that a distributed data store can provide at most two of three guarantees: consistency, availability, and partition tolerance.', 15),
(5, 1, 'Retrieval Practice', 'A study technique where you actively recall information from memory rather than passively reviewing it, shown to produce stronger long-term retention.', 13)
ON CONFLICT DO NOTHING;

SELECT setval('glossary_terms_id_seq', GREATEST(5, (SELECT MAX(id) FROM glossary_terms)));

-- ---------------------------------------------------------------------------
-- Note events (lifecycle transitions for non-fleeting notes)
-- ---------------------------------------------------------------------------
INSERT INTO note_events (id, note_id, user_id, from_status, to_status, created_at) VALUES
-- Sleeping notes: created as fleeting, then moved to sleeping
(1,  5,  1, NULL,       'fleeting',  NOW() - INTERVAL '14 days'),
(2,  5,  1, 'fleeting', 'sleeping',  NOW() - INTERVAL '12 days'),
(3,  6,  1, NULL,       'fleeting',  NOW() - INTERVAL '10 days'),
(4,  6,  1, 'fleeting', 'sleeping',  NOW() - INTERVAL '8 days'),
(5,  7,  1, NULL,       'fleeting',  NOW() - INTERVAL '7 days'),
(6,  7,  1, 'fleeting', 'sleeping',  NOW() - INTERVAL '5 days'),

-- Active notes: fleeting > sleeping > active
(7,  8,  1, NULL,       'fleeting',  NOW() - INTERVAL '30 days'),
(8,  8,  1, 'fleeting', 'sleeping',  NOW() - INTERVAL '25 days'),
(9,  8,  1, 'sleeping', 'active',    NOW() - INTERVAL '20 days'),
(10, 12, 1, NULL,       'fleeting',  NOW() - INTERVAL '21 days'),
(11, 12, 1, 'fleeting', 'sleeping',  NOW() - INTERVAL '18 days'),
(12, 12, 1, 'sleeping', 'active',    NOW() - INTERVAL '15 days'),
(13, 13, 1, NULL,       'fleeting',  NOW() - INTERVAL '20 days'),
(14, 13, 1, 'fleeting', 'sleeping',  NOW() - INTERVAL '16 days'),
(15, 13, 1, 'sleeping', 'active',    NOW() - INTERVAL '12 days'),

-- Linked notes: fleeting > sleeping > active > linked
(16, 9,  1, NULL,       'fleeting',  NOW() - INTERVAL '28 days'),
(17, 9,  1, 'fleeting', 'sleeping',  NOW() - INTERVAL '24 days'),
(18, 9,  1, 'sleeping', 'active',    NOW() - INTERVAL '18 days'),
(19, 9,  1, 'active',   'linked',    NOW() - INTERVAL '10 days'),
(20, 10, 1, NULL,       'fleeting',  NOW() - INTERVAL '26 days'),
(21, 10, 1, 'fleeting', 'sleeping',  NOW() - INTERVAL '22 days'),
(22, 10, 1, 'sleeping', 'active',    NOW() - INTERVAL '16 days'),
(23, 10, 1, 'active',   'linked',    NOW() - INTERVAL '8 days'),
(24, 11, 1, NULL,       'fleeting',  NOW() - INTERVAL '14 days'),
(25, 11, 1, 'fleeting', 'sleeping',  NOW() - INTERVAL '12 days'),
(26, 11, 1, 'sleeping', 'active',    NOW() - INTERVAL '9 days'),
(27, 11, 1, 'active',   'linked',    NOW() - INTERVAL '5 days'),
(28, 14, 1, NULL,       'fleeting',  NOW() - INTERVAL '18 days'),
(29, 14, 1, 'fleeting', 'sleeping',  NOW() - INTERVAL '14 days'),
(30, 14, 1, 'sleeping', 'active',    NOW() - INTERVAL '10 days'),
(31, 14, 1, 'active',   'linked',    NOW() - INTERVAL '4 days'),
(32, 15, 1, NULL,       'fleeting',  NOW() - INTERVAL '22 days'),
(33, 15, 1, 'fleeting', 'sleeping',  NOW() - INTERVAL '18 days'),
(34, 15, 1, 'sleeping', 'active',    NOW() - INTERVAL '14 days'),
(35, 15, 1, 'active',   'linked',    NOW() - INTERVAL '6 days'),

-- Archived notes: went through lifecycle then archived
(36, 16, 1, NULL,       'fleeting',  NOW() - INTERVAL '60 days'),
(37, 16, 1, 'fleeting', 'sleeping',  NOW() - INTERVAL '55 days'),
(38, 16, 1, 'sleeping', 'active',    NOW() - INTERVAL '45 days'),
(39, 16, 1, 'active',   'archived',  NOW() - INTERVAL '7 days'),
(40, 17, 1, NULL,       'fleeting',  NOW() - INTERVAL '50 days'),
(41, 17, 1, 'fleeting', 'archived',  NOW() - INTERVAL '40 days')
ON CONFLICT DO NOTHING;

SELECT setval('note_events_id_seq', GREATEST(41, (SELECT MAX(id) FROM note_events)));

COMMIT;
