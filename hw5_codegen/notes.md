# Заметки
## Алгоритм действий
1. Обход всего кода, анализ функций и структур 
2. Запись всех функций, обертки для которых необходимо сгенерировать, в структуру (информация о endpoint'е, имя структуры, к которой принадлежит функция, имена аргументов структур и имена структур возвращаемых функцией)
3. Запись всех структур в структуру: Имена и типы полей, параметры валидации. 
4. Генерация функций валидации для всех известных аргументов функций
5. Генерация обработчиков:
   1. Проверка существования метода
   2. Проверка авторизации 
   3. Валидация входных данных 
   4. Вызов функции
   5. Возврат значения