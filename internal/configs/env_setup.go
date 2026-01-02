package configs

import (
	"bufio"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
)

type DBInfo struct {
	Host    string
	User    string
	Pass    string
	Name    string
	Port    string
	Prefix  string
	Socket  string
	MaxIdle string
	MaxOpen string
}

type AdminInfo struct {
	Id string
	Pw string
}

// 게시판 생성 시 기본값 정의
const (
	CREATE_BOARD_ADMIN       = 1
	CREATE_BOARD_TYPE        = 0 /* board */
	CREATE_BOARD_NAME        = "board name"
	CREATE_BOARD_INFO        = "description for this board"
	CREATE_BOARD_ROWS        = 15
	CREATE_BOARD_WIDTH       = 1000
	CREATE_BOARD_USE_CAT     = 1
	CREATE_BOARD_LV_LIST     = 0
	CREATE_BOARD_LV_VIEW     = 0
	CREATE_BOARD_LV_WRITE    = 1 /* 0 is not allowed */
	CREATE_BOARD_LV_COMMENT  = 1 /* 0 is not allowed */
	CREATE_BOARD_LV_DOWNLOAD = 1 /* 0 is not allowed */
	CREATE_BOARD_PT_VIEW     = 0
	CREATE_BOARD_PT_WRITE    = 5
	CREATE_BOARD_PT_COMMENT  = 2
	CREATE_BOARD_PT_DOWNLOAD = -10
)

// 게시판 타입 목록
const (
	BOARD_DEFAULT = 0
	BOARD_GALLERY = 1
	BOARD_BLOG    = 2
	BOARD_SHOP    = 3
)

// NUBO 백엔드 실행 시 설치 여부 검사 후 필요 시 설치 진행
func Install() bool {
	if isInstalled := isAlreadyInstalled(); isInstalled {
		return true
	}

	welcome()

	dbInfo := askDBInfo()
	if len(dbInfo.Pass) < 1 {
		return false
	}

	adminInfo := askAdminInfo()
	if len(adminInfo.Id) < 1 {
		return false
	}

	if isEnv := makeEnv(dbInfo, adminInfo); !isEnv {
		return false
	}

	dbNoName, _ := connWithoutName(dbInfo)
	defer dbNoName.Close()

	if isDB := createDatabase(dbNoName, dbInfo.Name); !isDB {
		fmt.Printf(" [createDatabase] Failed to create database: %s\n", dbInfo.Name)
		return false
	}

	db, _ := connWithName(dbInfo)
	defer db.Close()

	createTables(db, dbInfo)
	insertRows(db, dbInfo, adminInfo)

	return true
}

// 바이너리 실행 시 "update" 인자가 넘어오면 업데이트 진행
func Update(db *sql.DB, prefix string) {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println("⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯")
	fmt.Printf(" → Update from ~v1.0.5 to %s\n", yellow("v2.0.0"))

	if err := createTradeTable(db, prefix); err != nil {
		fmt.Printf("%s\n", red(err.Error()))
	}

	fmt.Printf(" → created a new table: %s\n", green("trade"))
	fmt.Println(` → Now NUBO (goapi) starts a backend service`)
	fmt.Println("⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯")
}

// .env 파일이 존재하는지 확인하기
func isAlreadyInstalled() bool {
	info, err := os.Stat(".env")
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// NUBO (TSBOARD) 설치 웰컴 메시지 보여주기
func welcome() {
	fmt.Print(`
.:::     .::.::     .::.:: .::       .::::     
.: .::   .::.::     .::.:    .::   .::    .::  
.:: .::  .::.::     .::.:     .::.::        .::
.::  .:: .::.::     .::.::: .:   .::        .::
.::   .: .::.::     .::.:     .::.::        .::
.::    .: ::.::     .::.:      .:  .::     .:: 
.::      .::  .:::::   .:::: .::     .::::     
                                                                               
Networked Utilities & Builtin Options | https://nubohub.org                                              
`)
}

// NUBO (TSBOARD)에서 DB정보 사용을 위한 정보 확인하기
func askDBInfo() DBInfo {
	dbInfo := DBInfo{}
	reader := bufio.NewReader(os.Stdin)
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println("")
	fmt.Println("⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯")
	fmt.Printf(" NUBO is %s.\n We will now proceed with the installation process.\n\n", red("not installed yet"))
	fmt.Printf(" Before installing NUBO, make sure that\n `%s` is already installed on your server.\n\n", yellow("libvips"))
	fmt.Printf(" During the installation process,\n you will need the connection details\n for a pre-installed `%s` on your server.\n", yellow("MySQL(Mariadb)"))
	fmt.Println("⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯")
	fmt.Println("")

	for {
		fmt.Print(" → Enter the hostname (default is localhost): ")
		host, _ := reader.ReadString('\n')
		host = strings.TrimSpace(host)
		if host == "" {
			host = "localhost"
		}

		fmt.Print(" → Enter the username (default is root): ")
		user, _ := reader.ReadString('\n')
		user = strings.TrimSpace(user)
		if user == "" {
			user = "root"
		}

		fmt.Print(" → Enter the password: ")
		pass, _ := reader.ReadString('\n')
		pass = strings.TrimSpace(pass)

		fmt.Print(" → Enter the dbname (default is tsboard): ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			name = "tsboard"
		}

		fmt.Print(" → Enter the prefix of tables (default is `tsb_`): ")
		prefix, _ := reader.ReadString('\n')
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			prefix = "tsb_"
		}

		fmt.Print(" → Enter the port number (default is 3306): ")
		port, _ := reader.ReadString('\n')
		port = strings.TrimSpace(port)
		if port == "" {
			port = "3306"
		}

		fmt.Print(" → Enter the number of max idle (default is 10): ")
		maxIdle, _ := reader.ReadString('\n')
		maxIdle = strings.TrimSpace(maxIdle)
		if maxIdle == "" {
			maxIdle = "10"
		}

		fmt.Print(" → Enter the number of max open (default is 10): ")
		maxOpen, _ := reader.ReadString('\n')
		maxOpen = strings.TrimSpace(maxOpen)
		if maxOpen == "" {
			maxOpen = "10"
		}

		fmt.Println("")
		fmt.Printf(" ✔︎ default path of mysqld.sock on Ubuntu is %s\n", yellow("/var/run/mysqld/mysqld.sock"))
		fmt.Printf(" ✔︎ default path of mysqld.sock on Mac is %s\n", yellow("/tmp/mysql.sock"))
		fmt.Printf(" ✔︎ Windows does not have this file, %s would be okay\n", yellow("keep empty"))
		fmt.Println("")
		fmt.Print(" → Enter the path of mysqld.sock: ")
		socket, _ := reader.ReadString('\n')
		socket = strings.TrimSpace(socket)

		fmt.Println("")
		fmt.Println("⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯")
		fmt.Printf(" %s Hostname        : %s\n", green("✔︎"), yellow(host))
		fmt.Printf(" %s Username        : %s\n", green("✔︎"), yellow(user))
		fmt.Printf(" %s Password        : %s\n", green("✔︎"), yellow(pass))
		fmt.Printf(" %s Database name   : %s\n", green("✔︎"), yellow(name))
		fmt.Printf(" %s Prefix of table : %s\n", green("✔︎"), yellow(prefix))
		fmt.Printf(" %s Port number     : %s\n", green("✔︎"), yellow(port))
		fmt.Printf(" %s Socket path     : %s\n", green("✔︎"), yellow(socket))
		fmt.Printf(" %s Max idle        : %s\n", green("✔︎"), yellow(maxIdle))
		fmt.Printf(" %s Max open        : %s\n", green("✔︎"), yellow(maxOpen))
		fmt.Println("⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯")

		fmt.Println("")
		fmt.Printf(" → Are you sure all the information you entered is correct? [%ses/%so/%suit]: ", green("Y"), yellow("N"), red("Q"))
		isCorrect, _ := reader.ReadString('\n')
		isCorrect = strings.TrimSpace(isCorrect)
		answer := strings.ToLower(isCorrect)

		if answer == "y" || answer == "yes" {
			dbInfo.Host = host
			dbInfo.User = user
			dbInfo.Pass = pass
			dbInfo.Name = name
			dbInfo.Prefix = prefix
			dbInfo.Port = port
			dbInfo.Port = port
			dbInfo.Socket = socket
			dbInfo.MaxIdle = maxIdle
			dbInfo.MaxOpen = maxOpen

			if isConn := testConnDB(dbInfo); !isConn {
				fmt.Printf(" %s The NUBO %s to the database with the information you provided.\n", red("🞬"), red("could not connect"))
				fmt.Printf(" %s Please try again.\n\n", red("🞬"))
				continue
			} else {
				break
			}
		} else if answer == "n" || answer == "no" {
			continue
		} else {
			fmt.Printf(" %s The NUBO will now exit. To install or reinstall the NUBO, please delete the %s file first and then run this binary again.\n", red("🞬"), yellow(".env"))
			return DBInfo{}
		}
	}
	return dbInfo
}

// DB 이름 없이 연결하고 db 포인터 반환
func connWithoutName(dbInfo DBInfo) (*sql.DB, error) {
	addr := fmt.Sprintf("tcp(%s:%s)", dbInfo.Host, dbInfo.Port)
	if len(dbInfo.Socket) > 0 {
		addr = fmt.Sprintf("unix(%s)", dbInfo.Socket)
	}
	dsn := fmt.Sprintf("%s:%s@%s/", dbInfo.User, dbInfo.Pass, addr)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// DB 이름으로 연결 및 db 포인터 반환
func connWithName(dbInfo DBInfo) (*sql.DB, error) {
	addr := fmt.Sprintf("tcp(%s:%s)", dbInfo.Host, dbInfo.Port)
	if len(dbInfo.Socket) > 0 {
		addr = fmt.Sprintf("unix(%s)", dbInfo.Socket)
	}
	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&loc=Local", dbInfo.User, dbInfo.Pass, addr, dbInfo.Name)

	red := color.New(color.FgRed).SprintFunc()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf(" [connWithName] %s\n", red(err.Error()))
		return nil, err
	}

	if err = db.Ping(); err != nil {
		fmt.Printf(" [connWithName] %s\n", red(err.Error()))
		return nil, err
	}
	return db, nil
}

// DB 연결 시험하기
func testConnDB(dbInfo DBInfo) bool {
	db, err := connWithoutName(dbInfo)
	if err != nil {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf(" [testConnDB] %s\n", red(err.Error()))
		return false
	}
	defer db.Close()
	return true
}

// 관리자 ID, PW 정보 입력받기
func askAdminInfo() AdminInfo {
	adminInfo := AdminInfo{}
	reader := bufio.NewReader(os.Stdin)
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	for {
		fmt.Println("")
		fmt.Print(" → Enter the admin's email (e.g. sirini@gmail.com): ")
		id, _ := reader.ReadString('\n')
		id = strings.TrimSpace(id)

		fmt.Print(" → Enter the password for admin: ")
		pw, _ := reader.ReadString('\n')
		pw = strings.TrimSpace(pw)

		fmt.Println("")
		fmt.Println("⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯")
		fmt.Printf(" %s Admin's email   : %s\n", green("✔︎"), yellow(id))
		fmt.Printf(" %s Password        : %s\n", green("✔︎"), yellow(pw))
		fmt.Println("⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯")

		fmt.Println("")
		fmt.Printf(" → Are you sure all the information you entered is correct? [%ses/%so/%suit]: ", green("Y"), yellow("N"), red("Q"))
		isCorrect, _ := reader.ReadString('\n')
		isCorrect = strings.TrimSpace(isCorrect)
		answer := strings.ToLower(isCorrect)

		if answer == "y" || answer == "yes" {
			adminInfo.Id = id
			adminInfo.Pw = pw
			break
		} else if answer == "n" || answer == "no" {
			continue
		} else {
			return AdminInfo{}
		}
	}
	return adminInfo
}

// .env 파일 생성하기
func makeEnv(dbInfo DBInfo, adminInfo AdminInfo) bool {
	sample, err := os.ReadFile("env.sample")
	if err != nil {
		return false
	}
	env := string(sample)
	env = strings.ReplaceAll(env, "#dbhost#", dbInfo.Host)
	env = strings.ReplaceAll(env, "#dbuser#", dbInfo.User)
	env = strings.ReplaceAll(env, "#dbpass#", dbInfo.Pass)
	env = strings.ReplaceAll(env, "#dbname#", dbInfo.Name)
	env = strings.ReplaceAll(env, "#dbprefix#", dbInfo.Prefix)
	env = strings.ReplaceAll(env, "#dbsock#", dbInfo.Socket)
	env = strings.ReplaceAll(env, "#dbmaxidle#", dbInfo.MaxIdle)
	env = strings.ReplaceAll(env, "#dbmaxopen#", dbInfo.MaxOpen)
	env = strings.ReplaceAll(env, "#jwtsecret#", uuid.New().String())
	env = strings.ReplaceAll(env, "#adminid#", adminInfo.Id)
	env = strings.ReplaceAll(env, "#adminpw#", adminInfo.Pw)

	err = os.WriteFile(".env", []byte(env), 0644)
	return err == nil
}

// 데이터베이스 생성하기
func createDatabase(db *sql.DB, dbName string) bool {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
	_, err := db.Exec(query)
	return err == nil
}

// 테이블들 생성하기
func createTables(db *sql.DB, dbInfo DBInfo) {
	createUserTable(db, dbInfo.Prefix)
	createUserTokenTable(db, dbInfo.Prefix)
	createUserPermissionTable(db, dbInfo.Prefix)
	createUserVerificationTable(db, dbInfo.Prefix)
	createUserAccessLogTable(db, dbInfo.Prefix)
	createUserBlackListTable(db, dbInfo.Prefix)
	createReportTable(db, dbInfo.Prefix)
	createChatTable(db, dbInfo.Prefix)
	createGroupTable(db, dbInfo.Prefix)
	createBoardTable(db, dbInfo.Prefix)
	createBoardCategoryTable(db, dbInfo.Prefix)
	createPointHistoryTable(db, dbInfo.Prefix)
	createPostTable(db, dbInfo.Prefix)
	createHashtagTable(db, dbInfo.Prefix)
	createPostHashtagTable(db, dbInfo.Prefix)
	createPostLikeTable(db, dbInfo.Prefix)
	createCommentTable(db, dbInfo.Prefix)
	createCommentLikeTable(db, dbInfo.Prefix)
	createFileTable(db, dbInfo.Prefix)
	createFileThumbnailTable(db, dbInfo.Prefix)
	createImageTable(db, dbInfo.Prefix)
	createNotificationTable(db, dbInfo.Prefix)
	createExifTable(db, dbInfo.Prefix)
	createImageDescriptionTable(db, dbInfo.Prefix)
	createTradeTable(db, dbInfo.Prefix)
}

// 기본 레코드들 추가하기
func insertRows(db *sql.DB, dbInfo DBInfo, adminInfo AdminInfo) {
	insertDefaultGroup(db, dbInfo.Prefix)
	insertDefaultAdmin(db, dbInfo.Prefix, adminInfo)
	insertDefaultBoard(db, dbInfo.Prefix)
	insertDefaultCategory(db, dbInfo.Prefix)
	insertDefaultGallery(db, dbInfo.Prefix)
	insertDefaultGalleryCategory(db, dbInfo.Prefix)
}

// user 테이블 생성
func createUserTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %suser (
	uid INT UNSIGNED NOT NULL auto_increment,
  id VARCHAR(100) NOT NULL DEFAULT '',
  name VARCHAR(30) NOT NULL DEFAULT '',
  password CHAR(64) NOT NULL DEFAULT '',
  profile VARCHAR(300) NOT NULL DEFAULT '',
  level TINYINT UNSIGNED NOT NULL DEFAULT 0,
  point INT UNSIGNED NOT NULL DEFAULT 0,
  signature VARCHAR(300) NOT NULL DEFAULT '',
  signup BIGINT UNSIGNED NOT NULL DEFAULT 0,
  signin BIGINT UNSIGNED NOT NULL DEFAULT 0,
  blocked TINYINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix)
	db.Exec(query)
}

// user_token 테이블 생성
func createUserTokenTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %suser_token (
  user_uid INT UNSIGNED NOT NULL DEFAULT 0,
  refresh CHAR(64) NOT NULL DEFAULT '',
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  KEY (user_uid),
  CONSTRAINT fk_ut FOREIGN KEY (user_uid) REFERENCES %suser(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix)
	db.Exec(query)
}

// user_permission 테이블 생성
func createUserPermissionTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %suser_permission (
  uid INT UNSIGNED NOT NULL auto_increment,
  user_uid INT UNSIGNED NOT NULL DEFAULT 0,
  write_post TINYINT UNSIGNED NOT NULL DEFAULT '1',
  write_comment TINYINT UNSIGNED NOT NULL DEFAULT '1',
  send_chat TINYINT UNSIGNED NOT NULL DEFAULT '1',
  send_report TINYINT UNSIGNED NOT NULL DEFAULT '1',
  PRIMARY KEY (uid),
  KEY (user_uid),
  CONSTRAINT fk_up FOREIGN KEY (user_uid) REFERENCES %suser(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix)
	db.Exec(query)
}

// user_verification 테이블 생성
func createUserVerificationTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %suser_verification (
  uid INT UNSIGNED NOT NULL auto_increment,
  email VARCHAR(100) NOT NULL DEFAULT '',
  code CHAR(6) NOT NULL DEFAULT '',
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix)
	db.Exec(query)
}

// user_access_log 테이블 생성
func createUserAccessLogTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %suser_access_log (
  uid INT UNSIGNED NOT NULL auto_increment,
  user_uid INT UNSIGNED NOT NULL DEFAULT 0,
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix)
	db.Exec(query)
}

// user_black_list 테이블 생성
func createUserBlackListTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %suser_black_list (
  user_uid INT UNSIGNED NOT NULL DEFAULT 0,
  black_uid INT UNSIGNED NOT NULL DEFAULT 0,
  KEY (user_uid),
  CONSTRAINT fk_ubl FOREIGN KEY (user_uid) REFERENCES %suser(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix)
	db.Exec(query)
}

// report 테이블 생성
func createReportTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sreport (
  uid INT UNSIGNED NOT NULL auto_increment,
  to_uid INT UNSIGNED NOT NULL DEFAULT 0,
  from_uid INT UNSIGNED NOT NULL DEFAULT 0,
  request VARCHAR(1000) NOT NULL DEFAULT '',
  response VARCHAR(1000) NOT NULL DEFAULT '',
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  solved TINYINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid),
  KEY (solved)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix)
	db.Exec(query)
}

// chat 테이블 생성
func createChatTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %schat (
  uid INT UNSIGNED NOT NULL auto_increment,
  to_uid INT UNSIGNED NOT NULL DEFAULT 0,
  from_uid INT UNSIGNED NOT NULL DEFAULT 0,
  message VARCHAR(1000) NOT NULL DEFAULT '',
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid),
  KEY (to_uid),
  KEY (from_uid),
  CONSTRAINT fk_ct FOREIGN KEY (to_uid) REFERENCES %suser(uid),
  CONSTRAINT fk_cf FOREIGN KEY (from_uid) REFERENCES %suser(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix)
	db.Exec(query)
}

// group 테이블 생성
func createGroupTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sgroup (
  uid INT UNSIGNED NOT NULL auto_increment,
  id VARCHAR(30) NOT NULL DEFAULT '',
  admin_uid INT UNSIGNED NOT NULL DEFAULT 0,
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix)
	db.Exec(query)
}

// board 테이블 생성
func createBoardTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sboard (
  uid INT UNSIGNED NOT NULL auto_increment,
  id VARCHAR(30) NOT NULL DEFAULT '',
  group_uid INT UNSIGNED NOT NULL DEFAULT 0,
  admin_uid INT UNSIGNED NOT NULL DEFAULT 0,
  type TINYINT NOT NULL DEFAULT 0,
  name VARCHAR(20) NOT NULL DEFAULT '',
  info VARCHAR(100) NOT NULL DEFAULT '',
  row_count TINYINT UNSIGNED NOT NULL DEFAULT '20',
  width INT UNSIGNED NOT NULL DEFAULT '1000',
  use_category TINYINT UNSIGNED NOT NULL DEFAULT 0,
  level_list TINYINT UNSIGNED NOT NULL DEFAULT 0,
  level_view TINYINT UNSIGNED NOT NULL DEFAULT 0,
  level_write TINYINT UNSIGNED NOT NULL DEFAULT 0,
  level_comment TINYINT UNSIGNED NOT NULL DEFAULT 0,
  level_download TINYINT UNSIGNED NOT NULL DEFAULT 0,
  point_view INT NOT NULL DEFAULT 0,
  point_write INT NOT NULL DEFAULT 0,
  point_comment INT NOT NULL DEFAULT 0,
  point_download INT NOT NULL DEFAULT 0,
  PRIMARY KEY (uid),
  CONSTRAINT fk_bg FOREIGN KEY (group_uid) REFERENCES %sgroup(uid),
  CONSTRAINT fk_ba FOREIGN KEY (admin_uid) REFERENCES %suser(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix)
	db.Exec(query)
}

// board_category 테이블 생성
func createBoardCategoryTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sboard_category (
  uid INT UNSIGNED NOT NULL auto_increment,
  board_uid INT UNSIGNED NOT NULL DEFAULT 0,
  name VARCHAR(30) NOT NULL DEFAULT '',
  PRIMARY KEY (uid),
  KEY (board_uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix)
	db.Exec(query)
}

// point_history 테이블 생성
func createPointHistoryTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %spoint_history (
  uid INT UNSIGNED NOT NULL auto_increment,
  user_uid INT UNSIGNED NOT NULL DEFAULT 0,
  board_uid INT UNSIGNED NOT NULL DEFAULT 0,
  action TINYINT UNSIGNED NOT NULL DEFAULT 0,
  point INT NOT NULL DEFAULT 0,
  PRIMARY KEY (uid),
  KEY (user_uid),
  CONSTRAINT fk_ph_u FOREIGN KEY (user_uid) REFERENCES %suser(uid),
  CONSTRAINT fk_ph_b FOREIGN KEY (board_uid) REFERENCES %sboard(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix)
	db.Exec(query)
}

// post 테이블 생성
func createPostTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %spost (
  uid INT UNSIGNED NOT NULL auto_increment,
  board_uid INT UNSIGNED NOT NULL DEFAULT 0,
  user_uid INT UNSIGNED NOT NULL DEFAULT 0,
  category_uid INT UNSIGNED NOT NULL DEFAULT 0,
  title VARCHAR(300) NOT NULL DEFAULT '',
  content TEXT,
  submitted BIGINT UNSIGNED NOT NULL DEFAULT 0,
  modified BIGINT UNSIGNED NOT NULL DEFAULT 0,
  hit INT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 0,
  PRIMARY KEY (uid),
  KEY (board_uid),
  KEY (user_uid),
  KEY (category_uid),
  KEY (submitted),
  KEY (hit),
  KEY (status),
  CONSTRAINT fk_pb FOREIGN KEY (board_uid) REFERENCES %sboard(uid),
  CONSTRAINT fk_pu FOREIGN KEY (user_uid) REFERENCES %suser(uid),
  CONSTRAINT fk_pc FOREIGN KEY (category_uid) REFERENCES %sboard_category(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix, prefix)
	db.Exec(query)
}

// hashtag 테이블 생성
func createHashtagTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %shashtag (
  uid INT UNSIGNED NOT NULL auto_increment,
  name VARCHAR(30) NOT NULL DEFAULT '',
  used INT UNSIGNED NOT NULL DEFAULT 0,
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix)
	db.Exec(query)
}

// post_hashtag 테이블 생성
func createPostHashtagTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %spost_hashtag (
  board_uid INT UNSIGNED NOT NULL DEFAULT 0,
  post_uid INT UNSIGNED NOT NULL DEFAULT 0,
  hashtag_uid INT UNSIGNED NOT NULL DEFAULT 0,
  KEY (board_uid),
  KEY (post_uid),
  KEY (hashtag_uid),
  CONSTRAINT fk_phb FOREIGN KEY (board_uid) REFERENCES %sboard(uid),
  CONSTRAINT fk_php FOREIGN KEY (post_uid) REFERENCES %spost(uid),
  CONSTRAINT fk_phh FOREIGN KEY (hashtag_uid) REFERENCES %shashtag(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix, prefix)
	db.Exec(query)
}

// post_like 테이블 생성
func createPostLikeTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %spost_like (
  board_uid INT UNSIGNED NOT NULL DEFAULT 0,
  post_uid INT UNSIGNED NOT NULL DEFAULT 0,
  user_uid INT UNSIGNED NOT NULL DEFAULT 0,
  liked TINYINT UNSIGNED NOT NULL DEFAULT 0,
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  KEY (post_uid),
  KEY (user_uid),
  KEY (liked),
  CONSTRAINT fk_plb FOREIGN KEY (board_uid) REFERENCES %sboard(uid),
  CONSTRAINT fk_plp FOREIGN KEY (post_uid) REFERENCES %spost(uid),
  CONSTRAINT fk_plu FOREIGN KEY (user_uid) REFERENCES %suser(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix, prefix)
	db.Exec(query)
}

// comment 테이블 생성
func createCommentTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %scomment (
  uid INT UNSIGNED NOT NULL auto_increment,
  reply_uid INT UNSIGNED NOT NULL DEFAULT 0,
  board_uid INT UNSIGNED NOT NULL DEFAULT 0,
  post_uid INT UNSIGNED NOT NULL DEFAULT 0,
  user_uid INT UNSIGNED NOT NULL DEFAULT 0,
  content VARCHAR(10000) NOT NULL DEFAULT '',
  submitted BIGINT UNSIGNED NOT NULL DEFAULT 0,
  modified BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 0,
  PRIMARY KEY (uid),
  KEY (reply_uid),
  KEY (board_uid),
  KEY (post_uid),
  KEY (user_uid),
  KEY (submitted),
  KEY (status),
  CONSTRAINT fk_cb FOREIGN KEY (board_uid) REFERENCES %sboard(uid),
  CONSTRAINT fk_cp FOREIGN KEY (post_uid) REFERENCES %spost(uid),
  CONSTRAINT fk_cu FOREIGN KEY (user_uid) REFERENCES %suser(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix, prefix)
	db.Exec(query)
}

// comment_like 테이블 생성
func createCommentLikeTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %scomment_like (
  board_uid INT UNSIGNED NOT NULL DEFAULT 0,
  comment_uid INT UNSIGNED NOT NULL DEFAULT 0,
  user_uid INT UNSIGNED NOT NULL DEFAULT 0,
  liked TINYINT UNSIGNED NOT NULL DEFAULT 0,
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  KEY (comment_uid),
  KEY (user_uid),
  KEY (liked),
  CONSTRAINT fk_clb FOREIGN KEY (board_uid) REFERENCES %sboard(uid),
  CONSTRAINT fk_clc FOREIGN KEY (comment_uid) REFERENCES %scomment(uid),
  CONSTRAINT fk_clu FOREIGN KEY (user_uid) REFERENCES %suser(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix, prefix)
	db.Exec(query)
}

// file 테이블 생성
func createFileTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sfile (
  uid INT UNSIGNED NOT NULL auto_increment,
  board_uid INT UNSIGNED NOT NULL DEFAULT 0,
  post_uid INT UNSIGNED NOT NULL DEFAULT 0,
  name VARCHAR(100) NOT NULL DEFAULT '',
  path VARCHAR(300) NOT NULL DEFAULT '',
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid),
  KEY (post_uid),
  CONSTRAINT fk_fb FOREIGN KEY (board_uid) REFERENCES %sboard(uid),
  CONSTRAINT fk_fp FOREIGN KEY (post_uid) REFERENCES %spost(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix)
	db.Exec(query)
}

// file_thumbnail 테이블 생성
func createFileThumbnailTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sfile_thumbnail (
  uid INT UNSIGNED NOT NULL auto_increment,
  file_uid INT UNSIGNED NOT NULL DEFAULT 0,
  post_uid INT UNSIGNED NOT NULL DEFAULT 0,
  path VARCHAR(300) NOT NULL DEFAULT '',
  full_path VARCHAR(300) NOT NULL DEFAULT '',
  PRIMARY KEY (uid),
  KEY (file_uid),
  KEY (post_uid),
  CONSTRAINT fk_ftf FOREIGN KEY (file_uid) REFERENCES %sfile(uid),
  CONSTRAINT fk_ftp FOREIGN KEY (post_uid) REFERENCES %spost(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix)
	db.Exec(query)
}

// image 테이블 생성
func createImageTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %simage (
  uid INT UNSIGNED NOT NULL auto_increment,
  board_uid INT UNSIGNED NOT NULL DEFAULT 0,
  user_uid INT UNSIGNED NOT NULL DEFAULT 0,
  path VARCHAR(300) NOT NULL DEFAULT '',
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid),
  KEY (user_uid),
  CONSTRAINT fk_ib FOREIGN KEY (board_uid) REFERENCES %sboard(uid),
  CONSTRAINT fk_iu FOREIGN KEY (user_uid) REFERENCES %suser(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix)
	db.Exec(query)
}

// notification 테이블 생성
func createNotificationTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %snotification (
  uid INT UNSIGNED NOT NULL auto_increment,
  to_uid INT UNSIGNED NOT NULL DEFAULT 0,
  from_uid INT UNSIGNED NOT NULL DEFAULT 0,
  type TINYINT UNSIGNED NOT NULL DEFAULT 0,
  post_uid INT UNSIGNED NOT NULL DEFAULT 0,
  comment_uid INT UNSIGNED NOT NULL DEFAULT 0,
  checked TINYINT UNSIGNED NOT NULL DEFAULT 0,
  timestamp BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid),
  KEY (to_uid),
  KEY (from_uid),
  KEY (post_uid),
  KEY (checked),
  CONSTRAINT fk_nt FOREIGN KEY (to_uid) REFERENCES %suser(uid),
  CONSTRAINT fk_nf FOREIGN KEY (from_uid) REFERENCES %sboard(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix)
	db.Exec(query)
}

// exif 테이블 생성
func createExifTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sexif (
  uid INT UNSIGNED NOT NULL auto_increment,
  file_uid INT UNSIGNED NOT NULL DEFAULT 0,
  post_uid INT UNSIGNED NOT NULL DEFAULT 0,
  make VARCHAR(20) NOT NULL DEFAULT '',
  model VARCHAR(20) NOT NULL DEFAULT '',
  aperture INT UNSIGNED NOT NULL DEFAULT 0,
  iso INT UNSIGNED NOT NULL DEFAULT 0,
  focal_length INT UNSIGNED NOT NULL DEFAULT 0,
  exposure INT UNSIGNED NOT NULL DEFAULT 0,
  width INT UNSIGNED NOT NULL DEFAULT 0,
  height INT UNSIGNED NOT NULL DEFAULT 0,
  date BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (uid),
  KEY (file_uid),
  KEY (post_uid),
  CONSTRAINT fk_ef FOREIGN KEY (file_uid) REFERENCES %sfile(uid),
  CONSTRAINT fk_ep FOREIGN KEY (post_uid) REFERENCES %spost(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix)
	db.Exec(query)
}

// image_description 테이블 생성
func createImageDescriptionTable(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %simage_description (
  uid INT UNSIGNED NOT NULL auto_increment,
  file_uid INT UNSIGNED NOT NULL DEFAULT 0,
  post_uid INT UNSIGNED NOT NULL DEFAULT 0,
  description VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (uid),
  KEY (file_uid),
  KEY (post_uid),
  CONSTRAINT fk_idf FOREIGN KEY (file_uid) REFERENCES %sfile(uid),
  CONSTRAINT fk_idp FOREIGN KEY (post_uid) REFERENCES %spost(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix, prefix)
	db.Exec(query)
}

// trade 테이블 생성 (v1.0.4)
func createTradeTable(db *sql.DB, prefix string) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %strade (
	uid INT UNSIGNED NOT NULL auto_increment,
	post_uid INT UNSIGNED NOT NULL DEFAULT 0,
	brand VARCHAR(100) NOT NULL DEFAULT '',
	category TINYINT UNSIGNED NOT NULL DEFAULT 0,
	price INT UNSIGNED NOT NULL DEFAULT 0,
	product_condition TINYINT UNSIGNED NOT NULL DEFAULT 0,
	location VARCHAR(100) NOT NULL DEFAULT '',
	shipping_type TINYINT UNSIGNED NOT NULL DEFAULT 0,
	status TINYINT UNSIGNED NOT NULL DEFAULT 0,
	completed BIGINT UNSIGNED NOT NULL DEFAULT 0,
	PRIMARY KEY (uid),
	KEY (post_uid),
	KEY (status),
	CONSTRAINT fk_tpp FOREIGN KEY (post_uid) REFERENCES %spost(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci`, prefix, prefix)
	_, err := db.Exec(query)
	return err
}

// 기본 그룹 생성
func insertDefaultGroup(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`INSERT INTO %sgroup (id, admin_uid, timestamp) VALUES (?, ?, ?)`, prefix)
	db.Exec(query, "boards", 1, time.Now().UnixMilli())
}

// 기본 관리자 생성
func insertDefaultAdmin(db *sql.DB, prefix string, adminInfo AdminInfo) {
	hash := sha256.New()
	hash.Write([]byte(adminInfo.Pw))
	hashBytes := hash.Sum(nil)
	hashed := hex.EncodeToString(hashBytes)

	query := fmt.Sprintf(`INSERT INTO %suser (
		id, name, password, profile, level, point, signature, signup, signin, blocked
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, prefix)
	db.Exec(query, adminInfo.Id, "Admin", hashed, "", 9, 1000, "", time.Now().UnixMilli(), 0, 0)
}

// 기본 게시판 생성
func insertDefaultBoard(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`INSERT INTO %sboard (
  id, group_uid, admin_uid, type, name, info, row_count, width, use_category,
  level_list, level_view, level_write, level_comment, level_download,
  point_view, point_write, point_comment, point_download
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?,
  ?, ?, ?, ?
)`, prefix)
	db.Exec(query,
		"free",
		1,
		1,
		BOARD_DEFAULT,
		"free",
		"write everything you want",
		CREATE_BOARD_ROWS,
		CREATE_BOARD_WIDTH,
		CREATE_BOARD_USE_CAT,
		CREATE_BOARD_LV_LIST,
		CREATE_BOARD_LV_VIEW,
		CREATE_BOARD_LV_WRITE,
		CREATE_BOARD_LV_COMMENT,
		CREATE_BOARD_LV_DOWNLOAD,
		CREATE_BOARD_PT_VIEW,
		CREATE_BOARD_PT_WRITE,
		CREATE_BOARD_PT_COMMENT,
		CREATE_BOARD_PT_DOWNLOAD,
	)
}

// 기본 분류들 생성
func insertDefaultCategory(db *sql.DB, prefix string) {
	query := fmt.Sprintf("INSERT INTO %sboard_category (board_uid, name) VALUES (?, ?)", prefix)
	db.Exec(query, 1, "open")

	query = fmt.Sprintf("INSERT INTO %sboard_category (board_uid, name) VALUES (?, ?)", prefix)
	db.Exec(query, 1, "qna")

	query = fmt.Sprintf("INSERT INTO %sboard_category (board_uid, name) VALUES (?, ?)", prefix)
	db.Exec(query, 1, "news")
}

// 기본 갤러리 생성
func insertDefaultGallery(db *sql.DB, prefix string) {
	query := fmt.Sprintf(`INSERT INTO %sboard (
  id, group_uid, admin_uid, type, name, info, row_count, width, use_category,
  level_list, level_view, level_write, level_comment, level_download,
  point_view, point_write, point_comment, point_download
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?,
  ?, ?, ?, ?
)`, prefix)
	db.Exec(query,
		"photo",
		1,
		1,
		BOARD_GALLERY,
		"gallery",
		"home of photographers",
		CREATE_BOARD_ROWS,
		CREATE_BOARD_WIDTH,
		CREATE_BOARD_USE_CAT,
		CREATE_BOARD_LV_LIST,
		CREATE_BOARD_LV_VIEW,
		CREATE_BOARD_LV_WRITE,
		CREATE_BOARD_LV_COMMENT,
		CREATE_BOARD_LV_DOWNLOAD,
		CREATE_BOARD_PT_VIEW,
		CREATE_BOARD_PT_WRITE,
		CREATE_BOARD_PT_COMMENT,
		CREATE_BOARD_PT_DOWNLOAD,
	)
}

// 기본 갤러리의 분류들 생성
func insertDefaultGalleryCategory(db *sql.DB, prefix string) {
	query := fmt.Sprintf("INSERT INTO %sboard_category (board_uid, name) VALUES (?, ?)", prefix)
	db.Exec(query, 2, "daily")

	query = fmt.Sprintf("INSERT INTO %sboard_category (board_uid, name) VALUES (?, ?)", prefix)
	db.Exec(query, 2, "landscape")

	query = fmt.Sprintf("INSERT INTO %sboard_category (board_uid, name) VALUES (?, ?)", prefix)
	db.Exec(query, 2, "portrait")
}
