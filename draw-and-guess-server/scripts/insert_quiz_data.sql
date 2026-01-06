-- Insert OX Quiz Data
INSERT INTO ox_quizzes (id, category, difficulty, question, answer, explanation, usage_count, is_active, created_at, updated_at) VALUES
(1, 'Science', 'easy', 'The Earth is flat.', false, 'The Earth is approximately spherical in shape.', 0, true, NOW(), NOW()),
(2, 'History', 'easy', 'The Great Wall of China was built to keep out invaders.', true, 'The Great Wall was primarily built as a defense system against invasions.', 0, true, NOW(), NOW()),
(3, 'Geography', 'easy', 'Mount Everest is the tallest mountain in the world.', true, 'Mount Everest stands at 8,849 meters (29,032 feet) above sea level.', 0, true, NOW(), NOW()),
(4, 'Science', 'medium', 'Water boils at 100 degrees Celsius at sea level.', true, 'At standard atmospheric pressure (sea level), water boils at 100°C or 212°F.', 0, true, NOW(), NOW()),
(5, 'Biology', 'easy', 'Humans have five senses.', true, 'The traditional five senses are sight, hearing, touch, taste, and smell.', 0, true, NOW(), NOW()),
(6, 'History', 'medium', 'The United States declared independence in 1776.', true, 'The Declaration of Independence was signed on July 4, 1776.', 0, true, NOW(), NOW()),
(7, 'Science', 'easy', 'The Sun revolves around the Earth.', false, 'The Earth revolves around the Sun. This is called heliocentric model.', 0, true, NOW(), NOW()),
(8, 'Math', 'easy', 'A triangle has four sides.', false, 'A triangle has three sides. A shape with four sides is called a quadrilateral.', 0, true, NOW(), NOW()),
(9, 'Technology', 'easy', 'HTML stands for HyperText Markup Language.', true, 'HTML is the standard markup language for creating web pages.', 0, true, NOW(), NOW()),
(10, 'Geography', 'easy', 'Africa is the largest continent by area.', false, 'Asia is the largest continent by both area and population. Africa is the second largest.', 0, true, NOW(), NOW()),
(11, 'Biology', 'medium', 'Sharks are mammals.', false, 'Sharks are fish, not mammals. They breathe through gills and are cold-blooded.', 0, true, NOW(), NOW()),
(12, 'Physics', 'medium', 'Light travels faster than sound.', true, 'Light travels at approximately 299,792 km/s, while sound travels at about 343 m/s in air.', 0, true, NOW(), NOW()),
(13, 'History', 'easy', 'World War II ended in 1945.', true, 'World War II officially ended on September 2, 1945, with Japan''s surrender.', 0, true, NOW(), NOW()),
(14, 'Science', 'easy', 'Diamonds are made of carbon.', true, 'Diamonds are composed entirely of carbon atoms arranged in a crystal structure.', 0, true, NOW(), NOW()),
(15, 'Geography', 'medium', 'The Amazon River is longer than the Nile River.', false, 'The Nile River is generally considered the longest river in the world at about 6,650 km.', 0, true, NOW(), NOW());

-- Insert General/QA Quiz Data
INSERT INTO qa_quizzes (id, category, difficulty, question, answer, usage_count, is_active, created_at, updated_at) VALUES
(1, 'General Knowledge', 'easy', 'What is the capital of France?', 'Paris', 0, true, NOW(), NOW()),
(2, 'Science', 'easy', 'What is H2O commonly known as?', 'Water', 0, true, NOW(), NOW()),
(3, 'History', 'medium', 'Who painted the Mona Lisa?', 'Leonardo da Vinci', 0, true, NOW(), NOW()),
(4, 'Geography', 'easy', 'Which continent is known as the "Dark Continent"?', 'Africa', 0, true, NOW(), NOW()),
(5, 'Science', 'medium', 'What is the chemical symbol for gold?', 'Au', 0, true, NOW(), NOW()),
(6, 'Sports', 'easy', 'How many players are on a soccer team?', '11', 0, true, NOW(), NOW()),
(7, 'Technology', 'easy', 'What does CPU stand for?', 'Central Processing Unit', 0, true, NOW(), NOW()),
(8, 'Literature', 'medium', 'Who wrote "Romeo and Juliet"?', 'William Shakespeare', 0, true, NOW(), NOW()),
(9, 'Math', 'easy', 'What is the result of 7 x 8?', '56', 0, true, NOW(), NOW()),
(10, 'Science', 'easy', 'What planet is known as the Red Planet?', 'Mars', 0, true, NOW(), NOW()),
(11, 'History', 'hard', 'In which year did Christopher Columbus discover America?', '1492', 0, true, NOW(), NOW()),
(12, 'Geography', 'medium', 'What is the smallest country in the world?', 'Vatican City', 0, true, NOW(), NOW()),
(13, 'Music', 'easy', 'How many keys does a standard piano have?', '88', 0, true, NOW(), NOW()),
(14, 'Science', 'medium', 'What is the speed of light in vacuum (approximately)?', '300000 km/s', 0, true, NOW(), NOW()),
(15, 'General Knowledge', 'easy', 'What is the largest ocean on Earth?', 'Pacific Ocean', 0, true, NOW(), NOW());

-- Verify insertion
SELECT 'OX Quizzes' as type, COUNT(*) as count FROM ox_quizzes WHERE is_active = true
UNION ALL
SELECT 'QA Quizzes' as type, COUNT(*) as count FROM qa_quizzes WHERE is_active = true;
