-- 000046 already owns key=task_lifecycle. This slice only publishes a version
-- when none exists, so rolling back must not drop a designer-published graph
-- or the 000046 template.
SELECT 1;
