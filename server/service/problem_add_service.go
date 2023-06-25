package service

import (
	"fmt"
	"os"
	"singo/model"
	"singo/serializer"
)

type ProblemAddService struct {
	CreatorID uint   `form:"creator_id" json:"creator_id"`
	Title     string `form:"title" json:"title" binding:"required"`
	// $MemoLimit MB
	MemoLimit int64 `form:"memo_limit" json:"memo_limit" binding:"required"`
	// $TimeLimit ms
	TimeLimit int64  `form:"time_limit" json:"time_limit" binding:"required"`
	Text      string `form:"text" json:"text" binding:"required"`
}

func (service *ProblemAddService) valid() *serializer.Response {
	if service.MemoLimit < 16 { // < 16MB
		return &serializer.Response{
			Code: 40001,
			Msg:  "内存设置过小，至少为 16 MB",
		}
	}

	if service.TimeLimit < 500 { // < 500ms
		return &serializer.Response{
			Code: 40001,
			Msg:  "时间设置过小，至少为 500ms",
		}
	}

	return nil
}

func (service *ProblemAddService) Add() serializer.Response {
	if err := service.valid(); err != nil {
		return *err
	}

	problem := model.Problem{
		CreatorID: service.CreatorID,
		Title:     service.Title,
		MemoLimit: service.MemoLimit,
		TimeLimit: service.TimeLimit,
		Path:      "",
	}

	if err := model.DB.Create(&problem).Error; err != nil {
		return serializer.ParamErr("添加题目失败", err)
	}

	os.Mkdir("problems", os.ModePerm)
	path := fmt.Sprintf("problems/P%v", problem.ID)
	problem.Path = path
	if err := os.Mkdir(path, os.ModePerm); err != nil {
		return serializer.Err(
			serializer.CodeFileSystemError,
			fmt.Sprintf("%s 创建失败", path),
			err,
		)
	}
	textPath := path + "/text.md"
	fo, err := os.Create(textPath)
	if err != nil {
		return serializer.Err(
			serializer.CodeFileSystemError,
			fmt.Sprintf("%s 打开失败", textPath),
			err,
		)
	}
	defer fo.Close()
	if _, err := fo.Write([]byte(service.Text)); err != nil {
		return serializer.Err(
			serializer.CodeFileSystemError,
			fmt.Sprintf("%s 写入失败", textPath),
			err,
		)
	}

	if err := model.DB.Save(&problem).Error; err != nil {
		return serializer.ParamErr("添加题目失败", err)
	}

	return serializer.BuildProblemResponse(problem)
}
