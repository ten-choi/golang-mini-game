# Check MongoDB QA quizzes
$mongoQuery = @'
use draw_and_guess_db
db.qa_quizzes.countDocuments({is_active: true})
db.qa_quizzes.find({is_active: true}).limit(3)
'@

Write-Host "========== Checking QA Quizzes in MongoDB ==========" -ForegroundColor Cyan
mongosh "mongodb://admin:password@10.33.255.58:30017/draw_and_guess_db" --quiet --eval "db.qa_quizzes.countDocuments({is_active: true})"
Write-Host ""
mongosh "mongodb://admin:password@10.33.255.58:30017/draw_and_guess_db" --quiet --eval "db.qa_quizzes.find({is_active: true}).limit(3).toArray()"
