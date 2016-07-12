-- +migrate Up
-- password: $2y$05$yK291Unwid3erFGGlV29P.zmxUzwZLFXIgbflEmoRkxJGovE4OmW2

-- users
INSERT INTO users(id,username,email,password) VALUES(1,"johndoe","john.doe@gonews.acme","$2y$05$yK291Unwid3erFGGlV29P.zmxUzwZLFXIgbflEmoRkxJGovE4OmW2");
INSERT INTO users(id,username,email,password) VALUES(2,"janedoe","jane.doe@gonews.acme","$2y$05$yK291Unwid3erFGGlV29P.zmxUzwZLFXIgbflEmoRkxJGovE4OmW2");
INSERT INTO users(id,username,email,password) VALUES(3,"jackdoe","jack.doe@gonews.acme","$2y$05$yK291Unwid3erFGGlV29P.zmxUzwZLFXIgbflEmoRkxJGovE4OmW2");
INSERT INTO users(id,username,email,password) VALUES(4,"jefinerdoe","jenifer.doe@gonews.acme","$2y$05$yK291Unwid3erFGGlV29P.zmxUzwZLFXIgbflEmoRkxJGovE4OmW2");
INSERT INTO users(id,username,email,password) VALUES(5,"helenadoe","helena.doe@gonews.acme","$2y$05$yK291Unwid3erFGGlV29P.zmxUzwZLFXIgbflEmoRkxJGovE4OmW2");

-- threads
INSERT INTO threads(id,title,url,author_id) VALUES(1,"A new computer language","http://computer-language.acme/example.html",1);
INSERT INTO threads(id,title,url,author_id) VALUES(2,"The Acme MVC framework","http://mvc.acme/introduction",2);
INSERT INTO threads(id,title,url,author_id) VALUES(3,"Furtif, A Scalabe Blockchain Database","http://furtif.acme/blog?id=10",3);
INSERT INTO threads(id,title,url,author_id) VALUES(4,"Parsing PDF in Joom with Kana","http://blog.kana.acme/tutorials/parsing-pdf-in-joom",4);
INSERT INTO threads(id,title,url,author_id) VALUES(5,"Querify – An open-source Query Language","http://querify.acme/documentation/#querify",5);
INSERT INTO threads(id,title,url,author_id) VALUES(6,"JetSet, a professional Javascript and Typescript IDE","http://jetset-ide.acme/presendation.html",1);
INSERT INTO threads(id,title,url,author_id) VALUES(7,"New York: The Silicon Valley of Fooding","https://hipsters.acme/article/3494949",2);
INSERT INTO threads(id,title,url,author_id) VALUES(8,"Professor Jack Michael: The Secret of Our Success","https://hipsters.acme/article/394491",4);
INSERT INTO threads(id,title,url,author_id) VALUES(9,"The Difference Between New York, Washington DC, and the Seattle","https://hipsters.acme/article/94844",3);

-- thread_votes
INSERT INTO thread_votes(thread_id,author_id,score) VALUES(2,3,1);
INSERT INTO thread_votes(thread_id,author_id,score) VALUES(2,1,1);
INSERT INTO thread_votes(thread_id,author_id,score) VALUES(2,4,1);
INSERT INTO thread_votes(thread_id,author_id,score) VALUES(1,5,1);
INSERT INTO thread_votes(thread_id,author_id,score) VALUES(1,4,1);
INSERT INTO thread_votes(thread_id,author_id,score) VALUES(4,2,1);

-- comment_votes
INSERT INTO comment_votes(id,comment_id,author_id,score) VALUES(1,1,2,1);
INSERT INTO comment_votes(id,comment_id,author_id,score) VALUES(2,1,3,1);
INSERT INTO comment_votes(id,comment_id,author_id,score) VALUES(3,1,4,1);
INSERT INTO comment_votes(id,comment_id,author_id,score) VALUES(4,2,1,1);
INSERT INTO comment_votes(id,comment_id,author_id,score) VALUES(5,2,3,1);
INSERT INTO comment_votes(id,comment_id,author_id,score) VALUES(6,2,4,-1);

-- comments
INSERT INTO comments(id,thread_id,author_id,content) VALUES(1,1,2,"Thanks, it looks great");
INSERT INTO comments(id,thread_id,author_id,content) VALUES(2,1,1,"Hi folks, here is my new programming language!");
INSERT INTO comments(id,thread_id,author_id,content) VALUES(3,1,3,"Is it as fast as Java?");
INSERT INTO comments(id,thread_id,author_id,content) VALUES(4,2,4,"How does it compare to AngularJS?");
INSERT INTO comments(id,thread_id,author_id,content) VALUES(5,2,1,"But is it webscale ? /s");
INSERT INTO comments(id,thread_id,author_id,content) VALUES(6,4,4,"Here is a sample code:\r\k = tkana.New(@file('myfile'))\r\tk.Build('PDF')\r");
INSERT INTO comments(id,thread_id,author_id,content) VALUES(7,5,2,"How does it compare to SQL?");
INSERT INTO comments(id,thread_id,author_id,content) VALUES(8,5,4,"What is the license?");

-- child comments
INSERT INTO comments(id,thread_id,author_id,content,parent_id) VALUES(9,5,5,"It is easier to learn than SQL",7);
INSERT INTO comments(id,thread_id,author_id,content,parent_id) VALUES(10,5,5,"GPL-3.0 for non commercial use, there is also a commercial license.",8);
INSERT INTO comments(id,thread_id,author_id,content,parent_id) VALUES(11,5,4,"Nice thank you",10);

-- +migrate Down
DELETE FROM threads;
DELETE FROM comments;
DELETE FROM comment_votes;
DELETE FROM thread_votes;
DELETE FROM users;