package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/domain"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/sysinfo"
)

func newDoctorCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose OS, CPU, memory, disk, Docker Engine, and Compose prerequisites",
		RunE: func(cmd *cobra.Command, args []string) error {
			detector := sysinfo.NewDetector()
			uc := application.NewDiagnoseEnvironmentUseCase(detector)

			report, err := uc.Execute(cmd.Context())
			if err != nil {
				return domain.WrapError(domain.ExitGeneralFailure, "failed to run environment diagnostics", err)
			}

			if jsonOutput {
				data, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}
				cmd.Println(string(data))
			} else {
				printDoctorReport(cmd, report)
			}

			if report.HasFailures() {
				return domain.NewError(domain.ExitUnsupportedEnv, "diagnostic prerequisites not satisfied")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output diagnostics in JSON format")
	return cmd
}

func printDoctorReport(cmd *cobra.Command, report *domain.DiagnosticReport) {
	cmd.Println("\n=== VULNEX Pre-flight Environment Diagnostics ===")
	for _, c := range report.Checks {
		tag := "[✓ PASS]"
		if c.Status == domain.CheckWarn {
			tag = "[! WARN]"
		} else if c.Status == domain.CheckFail {
			tag = "[✗ FAIL]"
		}

		if c.Message != "" {
			cmd.Printf("  %-8s %-32s : %s\n", tag, c.Name, c.Message)
		} else {
			cmd.Printf("  %-8s %-32s\n", tag, c.Name)
		}

		if c.Remediation != "" && (c.Status == domain.CheckWarn || c.Status == domain.CheckFail) {
			cmd.Printf("           -> Action: %s\n", c.Remediation)
		}
	}
	cmd.Println("=================================================")
	if report.Status == domain.CheckPass {
		cmd.Println("All environment prerequisites are satisfied. Ready to install.")
	} else if report.Status == domain.CheckWarn {
		cmd.Println("Prerequisites satisfied with warnings. Please review recommendations.")
	} else {
		cmd.Println("One or more prerequisites failed. Please resolve above issues before installation.")
	}
}
