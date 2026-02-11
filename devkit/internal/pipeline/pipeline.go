package pipeline

import (
	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"golang.org/x/sync/errgroup"
)

type Task func() error

// Pipeline calls the functions in order and returns the first encountered error
func New(tasks ...Task) error {
	for _, task := range tasks {
		if task == nil {
			continue
		}
		if err := task(); err != nil {
			return err
		}
	}

	return nil
}

type conditional[T Task | AsyncTask] bool

func (conditional conditional[T]) Then(then T) T {
	if conditional {
		return then
	}
	return nil
}

func If[T Task | AsyncTask](condition bool) conditional[T] {
	return conditional[T](condition)
}

// ————————————————————————————————

type AsyncTask func(errs *errgroup.Group) error

func Async(tasks ...AsyncTask) Task {
	return func() error {
		var errs errgroup.Group

		for _, task := range tasks {
			if task == nil {
				continue
			}
			if err := task(&errs); err != nil {
				return err
			}
		}

		if err := errs.Wait(); err != nil {
			switch err := err.(type) {
			case *liberrors.DetailedError:
				return err
			default:
				return &liberrors.DetailedError{
					Label: "Task Error",
					Context: liberrors.DirContext{
						Path: drive.ToAbs("data"),
					},
					Details: err.Error(),
				}
			}
		}

		return nil
	}
}
